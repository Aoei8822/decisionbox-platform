package litellm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	goembedding "github.com/decisionbox-io/decisionbox/libs/go-common/embedding"
)

func TestCompatibleEmbedding(t *testing.T) {
	for _, prefix := range []string{"/v1", "/custom/api", ""} {
		t.Run(prefix, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != prefix+"/embeddings" {
					t.Errorf("path = %s", r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer endpoint-token" {
					t.Error("missing endpoint token")
				}
				var body embeddingRequest
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body.Model != "BAAI/bge-m3" || body.EncodingFormat != "float" || len(body.Input) != 2 {
					t.Errorf("body = %+v", body)
				}
				_, _ = w.Write([]byte(`{"data":[{"index":1,"embedding":[3,4]},{"index":0,"embedding":[1,2]}]}`))
			}))
			defer srv.Close()
			p, err := goembedding.NewProvider("openai-compatible", goembedding.ProviderConfig{"base_url": srv.URL + prefix + "/", "model": "BAAI/bge-m3", "dimensions": "2", "credentials_json": "endpoint-token"})
			if err != nil {
				t.Fatal(err)
			}
			v, err := p.Embed(context.Background(), []string{"ภาษาไทย", "English"})
			if err != nil {
				t.Fatal(err)
			}
			if v[0][0] != 1 || v[1][0] != 3 {
				t.Fatalf("wrong order: %v", v)
			}
		})
	}
}

func TestCompatibleInvalidURL(t *testing.T) {
	for _, base := range []string{"", "localhost:80", "ftp://host", "http://host/v1/embeddings", "http://user:pass@host", "http://host?key=secret", "http://host#fragment"} {
		if _, err := compatibleFactory(goembedding.ProviderConfig{"base_url": base}); err == nil {
			t.Errorf("accepted %q", base)
		}
	}
}

func TestCompatibleRejectsBadVectors(t *testing.T) {
	for _, body := range []string{
		`{"data":[{"index":0,"embedding":[]}]}`,
		`{"data":[{"index":0,"embedding":[1]}]}`,
		`{"data":[{"index":1,"embedding":[1,2]}]}`,
	} {
		t.Run(body, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
			defer srv.Close()
			p, err := compatibleFactory(goembedding.ProviderConfig{"base_url": srv.URL, "model": "m", "dimensions": "2", "auth_method": "none"})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := p.Embed(context.Background(), []string{"test"}); err == nil {
				t.Fatal("accepted malformed vector response")
			}
		})
	}
}

func TestCompatibleListOnlyAndNoAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("no-auth leaked a key")
		}
		if r.URL.Path != "/v1/models" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"my-custom-embedding"}]}`))
	}))
	defer srv.Close()
	p, err := compatibleFactory(goembedding.ProviderConfig{"base_url": srv.URL + "/v1", "auth_method": "none", "credentials_json": "stale-key"})
	if err != nil {
		t.Fatal(err)
	}
	models, err := p.(goembedding.ModelLister).ListModels(context.Background())
	if err != nil || len(models) != 1 || models[0].ID != "my-custom-embedding" {
		t.Fatalf("models=%v err=%v", models, err)
	}
	if _, err := compatibleFactory(goembedding.ProviderConfig{"base_url": srv.URL, "model": "m"}); err == nil {
		t.Fatal("model without dimensions accepted")
	}
}

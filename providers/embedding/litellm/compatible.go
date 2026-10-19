package litellm

import (
	"fmt"
	"net/url"
	"strings"

	goembedding "github.com/decisionbox-io/decisionbox/libs/go-common/embedding"
)

// Reuse the OpenAI wire implementation already linked into both API and agent.
// Keep the existing litellm ID and its URL normalization backward compatible.
func init() {
	goembedding.RegisterWithMeta("openai-compatible", compatibleFactory, goembedding.ProviderMeta{
		Name:        "OpenAI-compatible (Open-source models)",
		Description: "Connect directly to Hugging Face TEI or another OpenAI-compatible embedding server. No OpenAI account required.",
		ConfigFields: append([]goembedding.ConfigField{
			{Key: "base_url", Label: "API base URL", Required: true, Type: "string", Placeholder: "http://tei:80/v1", Description: "Exact API prefix including /v1 when required. Requests append /embeddings and /models. Do not enter the /embeddings endpoint itself."},
			{Key: "model", Label: "Model", Required: true, Type: "string", Placeholder: "BAAI/bge-m3", Description: "Enter the model ID served by your endpoint. The server must have this embedding model deployed; selecting a name does not download it."},
			{Key: "dimensions", Label: "Dimensions", Required: true, Type: "string", Description: "Expected vector size, e.g. 1024 for BGE-M3. Required when using a model."},
		}, tlsConfigFields()...),
		AuthMethods: []goembedding.AuthMethod{
			{ID: "api_key", Name: "API Key", Description: "The endpoint provider's token, not an OpenAI key.", Fields: []goembedding.ConfigField{
				{Key: "credentials_json", Label: "Endpoint API key", Required: true, Type: "credential"},
			}},
			{ID: "none", Name: "No authentication", Description: "For your local embedding server without authentication."},
		},
		Models: []goembedding.ModelInfo{
			{ID: "BAAI/bge-m3", Name: "BGE-M3 (multilingual, self-hosted)", Dimensions: 1024},
			{ID: "Qwen/Qwen3-Embedding-0.6B", Name: "Qwen3 Embedding 0.6B (self-hosted)", Dimensions: 1024},
		},
	})
}

func compatibleFactory(cfg goembedding.ProviderConfig) (goembedding.Provider, error) {
	base := strings.TrimRight(strings.TrimSpace(cfg["base_url"]), "/")
	u, err := url.Parse(base)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return nil, fmt.Errorf("openai-compatible embedding: base_url must be an HTTP(S) API prefix without credentials, query or fragment")
	}
	if strings.HasSuffix(u.Path, "/embeddings") {
		return nil, fmt.Errorf("openai-compatible embedding: use the API prefix, not the /embeddings endpoint")
	}
	prov, err := factory(cfg)
	if err != nil {
		return nil, err
	}
	p := prov.(*provider)
	if strings.TrimSpace(p.model) != "" && p.dims == 0 {
		return nil, fmt.Errorf("openai-compatible embedding: dimensions is required (e.g. 1024 for BGE-M3)")
	}
	p.baseURL = base // An explicit prefix also supports gateways without /v1.
	return p, nil
}

package ports

import designdomain "shipwright/internal/design/domain"

type Provider interface {
	Name() string
	Detect() designdomain.ProviderDetection
	Prepare(state any, request string) (*designdomain.ProviderResult, error)
	Generate(state any, request string) (*designdomain.ProviderResult, error)
	Publish(state any) (*designdomain.ProviderResult, error)
	Verify(state any) (*designdomain.ProviderResult, error)
	Report(state any) (*designdomain.ProviderResult, error)
}

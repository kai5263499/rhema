package v1

//go:generate oapi-codegen -package v1 -include-tags=v1 -generate types -o types.gen.go -old-config-style ../../api/v1.yaml
//go:generate oapi-codegen -package v1 -include-tags=v1 -generate spec,server -o server.gen.go -old-config-style ../../api/v1.yaml

//go:generate oapi-codegen -package client -include-tags=v1 -generate types -o ../../client/types.gen.go -old-config-style ../../api/v1.yaml
//go:generate oapi-codegen -package client -include-tags=v1 -generate client -o ../../client/client.gen.go -old-config-style ../../api/v1.yaml

const (
	HeaderXClientSDN = "X-SSL-CLIENT-S-DN"
)

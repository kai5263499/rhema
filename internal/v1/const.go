package v1

//go:generate oapi-codegen -package v1 -include-tags=strangle -generate client -o strangle.gen.go -old-config-style ../../api/v1.yaml
//go:generate oapi-codegen -package v1 -include-tags=v1 -generate types -o types.gen.go -old-config-style ../../api/v1.yaml
//go:generate oapi-codegen -package v1 -include-tags=v1 -generate spec,server -o server.gen.go -old-config-style ../../api/v1.yaml

//go:generate oapi-codegen -package agentkiosk -include-tags=v1 -generate types -o ../../rhemaclient/types.gen.go -old-config-style ../../api/v1.yaml
//go:generate oapi-codegen -package agentkiosk -include-tags=v1 -generate client -o ../../rhemaclient/client.gen.go -old-config-style ../../api/v1.yaml

//go:generate swag init -d ../../cmd/apiserver -o ../../docs

const (
	HeaderXClientSDN = "X-SSL-CLIENT-S-DN"
)

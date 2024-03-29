package service

import (
	"fmt"
	"github.com/specgen-io/specgen-golang/v2/goven/spec"
	"strings"
)

func formBodyTypeName(operation *spec.NamedOperation) string {
	parts := strings.Split(operation.Name.Source, fmt.Sprintf(`%s_`, operation.InApi.Name.Source))
	typeName := strings.Join(parts[1:], "")
	return typeName
}

func logFieldsName(operation *spec.NamedOperation) string {
	return fmt.Sprintf("log%s", operation.Name.PascalCase())
}

func serviceInterfaceTypeVar(api *spec.Api) string {
	return fmt.Sprintf(`%sService`, api.Name.Source)
}

func serviceApiNameVersioned(api *spec.Api) string {
	return fmt.Sprintf(`%sService%s`, api.Name.Source, api.InHttp.InVersion.Name.PascalCase())
}

func serviceApiPublicNameVersioned(api *spec.Api) string {
	return fmt.Sprintf(`%sService%s`, api.Name.PascalCase(), api.InHttp.InVersion.Name.PascalCase())
}

func callCheckContentType(logFieldsVar, expectedContentType, requestVar, responseVar string) string {
	return fmt.Sprintf(`contenttype.Check(%s, %s, %s, %s)`, logFieldsVar, expectedContentType, requestVar, responseVar)
}

func addRoutesMethodName(api *spec.Api) string {
	return fmt.Sprintf(`Add%sRoutes`, api.Name.PascalCase())
}

func genFmtSprintf(format string, args ...string) string {
	if len(args) > 0 {
		return fmt.Sprintf(`fmt.Sprintf("%s", %s)`, format, strings.Join(args, ", "))
	} else {
		return format
	}
}

func routingPackageAlias(version *spec.Version) string {
	if version.Name.Source != "" {
		return fmt.Sprintf(`%s`, version.Name.FlatCase())
	} else {
		return fmt.Sprintf(`root`)
	}
}

func apiPackageAlias(api *spec.Api) string {
	version := api.InHttp.InVersion.Name
	if version.Source != "" {
		return api.Name.CamelCase() + version.PascalCase()
	}
	return api.Name.CamelCase()
}

func serviceCall(serviceVar string, operation *spec.NamedOperation) string {
	params := []string{}
	if operation.BodyIs(spec.RequestBodyString) {
		params = append(params, "body")
	}
	if operation.BodyIs(spec.RequestBodyJson) {
		params = append(params, "&body")
	}
	if operation.BodyIs(spec.RequestBodyFormData) {
		for _, param := range operation.Body.FormData {
			params = append(params, param.Name.CamelCase())
		}
	}
	if operation.BodyIs(spec.RequestBodyFormUrlEncoded) {
		for _, param := range operation.Body.FormUrlEncoded {
			params = append(params, param.Name.CamelCase())
		}
	}
	for _, param := range operation.QueryParams {
		params = append(params, param.Name.CamelCase())
	}
	for _, param := range operation.HeaderParams {
		params = append(params, param.Name.CamelCase())
	}
	for _, param := range operation.Endpoint.UrlParams {
		params = append(params, param.Name.CamelCase())
	}

	return fmt.Sprintf(`%s.%s(%s)`, serviceVar, operation.Name.PascalCase(), strings.Join(params, ", "))
}

func operationHasParams(api *spec.Api) bool {
	for _, operation := range api.Operations {
		for _, param := range operation.QueryParams {
			if &param != nil {
				return true
			}
		}
		for _, param := range operation.HeaderParams {
			if &param != nil {
				return true
			}
		}
		for _, param := range operation.Endpoint.UrlParams {
			if &param != nil {
				return true
			}
		}
	}
	return false
}

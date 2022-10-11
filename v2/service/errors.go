package service

import (
	"fmt"
	"github.com/specgen-io/specgen-golang/v2/imports"
	"github.com/specgen-io/specgen-golang/v2/module"
	"github.com/specgen-io/specgen-golang/v2/writer"

	"github.com/specgen-io/specgen-golang/v2/goven/generator"
	"github.com/specgen-io/specgen-golang/v2/goven/spec"
	"github.com/specgen-io/specgen-golang/v2/types"
)

func generateErrors(module, errorsModelsModule, respondModule module.Module, responses *spec.Responses) *generator.CodeFile {
	w := writer.New(module, "responses.go")

	imports := imports.New()
	imports.AddAlias("github.com/sirupsen/logrus", "log")
	imports.Add("net/http")
	imports.AddAlias(errorsModelsModule.Package, types.ErrorsModelsPackage)
	imports.Add(respondModule.Package)
	imports.Write(w)

	w.EmptyLine()
	badRequest := responses.GetByStatusName(spec.HttpStatusBadRequest)
	w.Line(`func RespondBadRequest(logFields log.Fields, res http.ResponseWriter, error *%s) {`, types.GoType(&badRequest.Type.Definition))
	w.Line(`  log.WithFields(logFields).Warn(error.Message)`)
	generateResponseWriting(w.Indented(), `logFields`, badRequest, `error`)
	w.Line(`}`)

	w.EmptyLine()
	notFound := responses.GetByStatusName(spec.HttpStatusNotFound)
	w.Line(`func RespondNotFound(logFields log.Fields, res http.ResponseWriter, error *%s) {`, types.GoType(&notFound.Type.Definition))
	w.Line(`  log.WithFields(logFields).Warn(error.Message)`)
	generateResponseWriting(w.Indented(), `logFields`, notFound, `error`)
	w.Line(`}`)

	w.EmptyLine()
	internalServerError := responses.GetByStatusName(spec.HttpStatusInternalServerError)
	w.Line(`func RespondInternalServerError(logFields log.Fields, res http.ResponseWriter, error *%s) {`, types.GoType(&internalServerError.Type.Definition))
	w.Line(`  log.WithFields(logFields).Warn(error.Message)`)
	generateResponseWriting(w.Indented(), `logFields`, internalServerError, `error`)
	w.Line(`}`)

	return w.ToCodeFile()
}

func callCheckContentType(logFieldsVar, expectedContentType, requestVar, responseVar string) string {
	return fmt.Sprintf(`contenttype.Check(%s, %s, %s, %s)`, logFieldsVar, expectedContentType, requestVar, responseVar)
}

func respondNotFound(w generator.Writer, operation *spec.NamedOperation, message string) {
	specification := operation.InApi.InHttp.InVersion.InSpec
	badRequest := specification.HttpErrors.Responses.GetByStatusName(spec.HttpStatusNotFound)
	errorMessage := fmt.Sprintf(`%s{Message: %s}`, types.GoType(&badRequest.Type.Definition), message)
	w.Line(`httperrors.RespondNotFound(%s, res, &%s)`, logFieldsName(operation), errorMessage)
	w.Line(`return`)
}

func respondBadRequest(w generator.Writer, operation *spec.NamedOperation, location string, message string, params string) {
	specification := operation.InApi.InHttp.InVersion.InSpec
	badRequest := specification.HttpErrors.Responses.GetByStatusName(spec.HttpStatusBadRequest)
	errorMessage := fmt.Sprintf(`%s{Location: "%s", Message: %s, Errors: %s}`, types.GoType(&badRequest.Type.Definition), location, message, params)
	w.Line(`httperrors.RespondBadRequest(%s, res, &%s)`, logFieldsName(operation), errorMessage)
	w.Line(`return`)
}

func respondInternalServerError(w generator.Writer, operation *spec.NamedOperation, message string) {
	specification := operation.InApi.InHttp.InVersion.InSpec
	internalServerError := specification.HttpErrors.Responses.GetByStatusName(spec.HttpStatusInternalServerError)
	errorMessage := fmt.Sprintf(`%s{Message: %s}`, types.GoType(&internalServerError.Type.Definition), message)
	w.Line(`httperrors.RespondInternalServerError(%s, res, &%s)`, logFieldsName(operation), errorMessage)
	w.Line(`return`)
}

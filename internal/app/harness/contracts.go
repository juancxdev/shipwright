package harness

import contracts "shipwright/internal/contracts/application"

const ContractFile = contracts.ContractFile

type ContractEndpoint = contracts.ContractEndpoint
type ContractSchema = contracts.ContractSchema
type ContractSpec = contracts.ContractSpec
type ContractParseResult = contracts.ContractParseResult
type MockCompliance = contracts.MockCompliance
type MockEndpointCoverage = contracts.MockEndpointCoverage
type BackendCompliance = contracts.BackendCompliance
type EndpointMatch = contracts.EndpointMatch
type SchemaMatch = contracts.SchemaMatch
type ContractValidation = contracts.ContractValidation

func ParseContract(path string) *ContractParseResult { return contracts.ParseContract(path) }
func ContractExists() bool                           { return contracts.ContractExists() }
func FormatContractSpec(spec *ContractSpec) string   { return contracts.FormatContractSpec(spec) }
func CheckMockCompliance(spec *ContractSpec) *MockCompliance {
	return contracts.CheckMockCompliance(spec)
}
func CheckBackendCompliance(spec *ContractSpec) *BackendCompliance {
	return contracts.CheckBackendCompliance(spec)
}
func ValidateContract(path string) *ContractValidation { return contracts.ValidateContract(path) }
func GenerateFrontendTasks(spec *ContractSpec, projectName string) string {
	return contracts.GenerateFrontendTasks(spec, projectName)
}
func GenerateBackendTasks(spec *ContractSpec, projectName string) string {
	return contracts.GenerateBackendTasks(spec, projectName)
}
func GenerateContractTasks(spec *ContractSpec, projectName string) (feTasks, beTasks string) {
	return contracts.GenerateContractTasks(spec, projectName)
}

func FormatMockCompliance(mc *MockCompliance) string { return contracts.FormatMockCompliance(mc) }
func FormatBackendCompliance(bc *BackendCompliance) string {
	return contracts.FormatBackendCompliance(bc)
}

package harness

import (
	doctor "shipwright/internal/doctor/application"
	integrations "shipwright/internal/integrations/application"
	platform "shipwright/internal/platform/application"
)

const (
	DoctorSeverityOK      = doctor.DoctorSeverityOK
	DoctorSeverityInfo    = doctor.DoctorSeverityInfo
	DoctorSeverityWarning = doctor.DoctorSeverityWarning
	DoctorSeverityError   = doctor.DoctorSeverityError
)

type DoctorReport = doctor.DoctorReport
type DoctorCheck = doctor.DoctorCheck
type DoctorSummary = doctor.DoctorSummary

func RunDoctor(probe platform.SystemProbe) (*DoctorReport, error) {
	return doctor.RunDoctor(probe)
}

func RunDoctorWithHealth(probe platform.SystemProbe, healthProbe integrations.HealthProbe) (*DoctorReport, error) {
	return doctor.RunDoctorWithHealth(probe, healthProbe)
}

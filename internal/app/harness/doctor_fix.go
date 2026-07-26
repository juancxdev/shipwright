package harness

import (
	config "shipwright/internal/config/application"
	doctor "shipwright/internal/doctor/application"
	platform "shipwright/internal/platform/application"
)

type DoctorFixResult = doctor.DoctorFixResult

func ApplyDoctorFixes(probe platform.SystemProbe) (*DoctorFixResult, error) {
	return doctor.ApplyDoctorFixes(probe)
}

func RepairPortableConfig(cfg *config.PortableConfig) []string {
	return doctor.RepairPortableConfig(cfg)
}

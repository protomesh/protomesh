package graviflow

type FileFormat int

const (
	PKCS1_FileFormat FileFormat = iota
	PKCS8_FileFormat
	PKIX_FileFormat
	SEC1_FileFormat
)

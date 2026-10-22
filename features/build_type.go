package features

type TypeOfBuild string

const (
	BuildTypeUndefined TypeOfBuild = ""
	BuildTypeOSS       TypeOfBuild = "oss"
	BuildTypeInternal  TypeOfBuild = "internal"
)

package backup

const volumesDataFileName = "volumes-data.yml"

type volumeData struct {
	Id     string `yaml:"id"`
	Type   string `yaml:"type"`
	Target string `yaml:"target"`
}

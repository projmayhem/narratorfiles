package objecttype

type ObjectType int

const (
	Other ObjectType = iota
	Directory
	Audio
)

func (o ObjectType) String() string {
	switch o {
	case Directory:
		return "directory"
	case Audio:
		return "audio"
	case Other:
		return "other"
	default:
		return "unknown"
	}
}

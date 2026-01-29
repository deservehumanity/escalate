package persistance

type Serializable interface {
	Serialize() ([]byte, error)
}

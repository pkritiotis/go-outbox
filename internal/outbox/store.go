package outbox

type StoreType int

const (
	SQL StoreType = iota
)

type Store interface {
	Type() StoreType
	SaveTx(message Message, tx interface{}) error
	Update(message Message) error
	GetUnprocessed() ([]Message, error)
}

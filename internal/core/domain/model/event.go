package model

// ini adalah kontrak untuk semua event yang nantinya boleh dikirim oleh Producer[T]
// artinya: setiap tipe data yang ingin digunakan sebagai event harus punya method GetId() string
type Event interface {
	GetId() string
}

package siaencoding

import (
	"testing"
)

var (
	i32 int32   = 0x01020304
	u32 uint32  = 0x01020304
	f32 float32 = 0x01020304
	i64 int64   = 0x0102030405060708
	u64 uint64  = 0x0102030405060708
	f64 float64 = 0x0102030405060708
)

// TestEncoding checks that the Enc/Dec function pairs are proper inverses of each other.
func TestEncoding(t *testing.T) {
	if DecInt32(EncInt32(i32)) != i32 {
		t.Fatal("int32 encode/decode mismatch")
	}
	if DecUint32(EncUint32(u32)) != u32 {
		t.Fatal("uint32 encode/decode mismatch")
	}
	if DecFloat32(EncFloat32(f32)) != f32 {
		t.Fatal("float32 encode/decode mismatch")
	}
	if DecInt64(EncInt64(i64)) != i64 {
		t.Fatal("int64 encode/decode mismatch")
	}
	if DecUint64(EncUint64(u64)) != u64 {
		t.Fatal("uint64 encode/decode mismatch")
	}
	if DecFloat64(EncFloat64(f64)) != f64 {
		t.Fatal("float64 encode/decode mismatch")
	}
}

func BenchmarkEncoding(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DecInt32(EncInt32(i32))
		DecUint32(EncUint32(u32))
		DecFloat32(EncFloat32(f32))
		DecInt64(EncInt64(i64))
		DecUint64(EncUint64(u64))
		DecFloat64(EncFloat64(f64))
	}
}

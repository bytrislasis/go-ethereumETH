package satoshiturk

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"

	"context"
)

func CheckKeyExists(address common.Address) (bool, error) {
	// Redis bağlantısı oluştur
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // Redis şifresi, eğer varsa
		DB:       0,  // Seçilecek Redis veritabanı
	})

	// Adresi string'e dönüştür
	addressStr := address.Hex()

	// Redis üzerinde anahtarı ara
	keyExists, err := rdb.Exists(context.Background(), addressStr).Result()
	if err != nil {
		// Hata oluştu
		return false, err
	}

	if keyExists == 1 {
		// Anahtar mevcut
		fmt.Printf("Key '%s' exists in Redis\n", addressStr)
		return true, nil
	} else {
		// Anahtar mevcut değil
		fmt.Printf("Key '%s' does not exist in Redis\n", addressStr)
		return false, nil
	}
}

package satoshiturk

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/crypto"
)

func DerivePath(extKey *hdkeychain.ExtendedKey, path string) (*hdkeychain.ExtendedKey, error) {
	segments := strings.Split(path, "/")
	childKey := extKey

	for _, segment := range segments {
		index, err := strconv.ParseUint(segment, 10, 32)
		if err != nil {
			return nil, err
		}

		childKey, err = childKey.Child(uint32(index))
		if err != nil {
			return nil, err
		}
	}

	return childKey, nil
}

func Generate(start uint32, num uint32, publicKey string, maxProcs uint32) {

	runtime.GOMAXPROCS(int(maxProcs))

	extPubKeyStr := publicKey

	extKey, err := hdkeychain.NewKeyFromString(extPubKeyStr)
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis sunucusu adresi
		Password: "",               // Redis şifresi, eğer varsa
		DB:       0,                // Seçilecek Redis veritabanı
	})

	startIndex := start
	numAddresses := num
	numThreads := uint32(5000)

	addressesPerThread := numAddresses / numThreads

	startTime := time.Now()

	var wg sync.WaitGroup
	wg.Add(int(numThreads))

	for i := uint32(0); i < numThreads; i++ {
		go setAddressWorker(startIndex+i*addressesPerThread, addressesPerThread, extKey, rdb, &wg)
	}

	wg.Wait()

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)

	minutes := int(elapsedTime.Minutes())
	seconds := int(elapsedTime.Seconds()) % 60
	milliseconds := elapsedTime.Milliseconds() % 1000

	fmt.Printf("\nAdres ekleme süresi: %d dakika, %d saniye, %d milisaniye\n", minutes, seconds, milliseconds)

	/*	startTime = time.Now()

		wg.Add(int(numThreads))*/

	// Adresleri kontrol eden iş parçacıklarını başlat
	/*for i := uint32(0); i < numThreads; i++ {
		go checkAddressWorker(startIndex+i*addressesPerThread, addressesPerThread, extKey, rdb, &wg)
	}*/

	/*	wg.Wait()

		// Bitiş zamanını kaydetme ve süre hesaplama
		endTime = time.Now()
		elapsedTime = endTime.Sub(startTime)

		// Süreyi dakika, saniye ve milisaniye cinsinden yazdırma
		minutes = int(elapsedTime.Minutes())
		seconds = int(elapsedTime.Seconds()) % 60
		milliseconds = elapsedTime.Milliseconds() % 1000*/

	rdb.Close()

	/*fmt.Printf("\nAdres kontrol süresi: %d dakika, %d saniye, %d milisaniye\n", minutes, seconds, milliseconds)*/
}

func setAddressWorker(start, count uint32, extKey *hdkeychain.ExtendedKey, rdb *redis.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := start; i < start+count; i++ {
		path := fmt.Sprintf("0/%d", i)

		childKey, err := DerivePath(extKey, path)
		if err != nil {
			panic(err)
		}

		rawPubKey, err := childKey.ECPubKey()
		if err != nil {
			panic(err)
		}

		ethAddress := crypto.PubkeyToAddress(*rawPubKey.ToECDSA())

		err = SetAddress(rdb, ethAddress)
		if err != nil {
			panic(err)
		}
	}
}

func checkAddressWorker(start, count uint32, extKey *hdkeychain.ExtendedKey, rdb *redis.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := start; i < start+count; i++ {
		path := fmt.Sprintf("0/%d", i)

		childKey, err := DerivePath(extKey, path)
		if err != nil {
			panic(err)
		}

		rawPubKey, err := childKey.ECPubKey()
		if err != nil {
			panic(err)
		}

		ethAddress := crypto.PubkeyToAddress(*rawPubKey.ToECDSA())

		exists, err := CheckKeyExists(rdb, ethAddress)
		if err != nil {
			panic(err)
		}

		if exists {
			fmt.Printf("Key '%s' exists in Redis\n", ethAddress.Hex())
		}
	}
}

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	badger "github.com/dgraph-io/badger/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/pubgo/tikdog/tikdog_watcher"
	"github.com/pubgo/xerror"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type WatcherFile struct {
	Path    string
	Hash    string
	Changed bool
	Synced  bool
}

func getBytes(data interface{}) []byte {
	dt, _ := jsoniter.Marshal(data)
	return dt
}

func getHash(path string) (hash string) {
	file, err := os.Open(path)
	if err == nil {
		h_ob := sha256.New()
		_, err := io.Copy(h_ob, file)
		if err == nil {
			hash := h_ob.Sum(nil)
			hashvalue := hex.EncodeToString(hash)
			return hashvalue
		} else {
			return "something wrong when use sha256 interface..."
		}
	} else {
		fmt.Printf("failed to open %s\n", path)
	}
	defer file.Close()
	return
}

// 本地文件加载
// 本地存储中，如果已经同步了，那么就不用同步了
//

func main() {
	tikdog_watcher.Start()

	var prefix = "watcher_file:"
	var ext = "drawio"
	var dir = os.ExpandEnv("${HOME}/Documents")
	fmt.Println(dir, ext)

	db, err := badger.Open(badger.DefaultOptions("demo"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	xerror.Panic(tikdog_watcher.AddRecursive(dir, func(event interface{}) error {
		switch event := event.(type) {
		case tikdog_watcher.Event:
			if tikdog_watcher.IsUpdateEvent(event) {
				key := []byte(prefix + event.Name)
				return xerror.Wrap(db.View(func(txn *badger.Txn) error {
					return xerror.Wrap(txn.Set(key, getBytes(&WatcherFile{
						Synced:  false,
						Changed: true,
						Path:    event.Name,
						Hash:    getHash(event.Name),
					})))
				}))
			}
		}
		return nil
	}))

	xerror.Panic(filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ext) {
			return nil
		}

		key := []byte(prefix + path)

		var val *WatcherFile
		xerror.Panic(db.View(func(txn *badger.Txn) error {
			itm, err := txn.Get(key)
			if err != badger.ErrKeyNotFound {
				return nil
			}

			xerror.Panic(err)
			xerror.Panic(itm.Value(func(_val []byte) error { return xerror.Wrap(jsoniter.Unmarshal(_val, val)) }))
			return nil
		}))

		hash := getHash(path)

		// 不存在或者修改了
		if val == nil || val.Hash != hash {
			//	oss put
			return nil
		}

		return xerror.Wrap(db.Update(func(txn *badger.Txn) error {
			return xerror.Wrap(txn.Set(key, getBytes(&WatcherFile{
				Synced:  false,
				Changed: true,
				Path:    path,
				Hash:    getHash(path),
			})))
		}))
	}))

	client, err := oss.New(
		os.Getenv("oss_endpoint"),
		os.Getenv("oss_ak"),
		os.Getenv("oss_sk"))
	xerror.Panic(err)

	kk := xerror.PanicErr(client.Bucket("kooksee")).(*oss.Bucket)

	for _, k := range xerror.PanicErr(kk.ListObjectsV2()).(oss.ListObjectsResultV2).Objects {
		fmt.Printf("%#v\n", k)
	}

	lsRes, err := client.ListBuckets()
	xerror.Panic(err)

	for _, bucket := range lsRes.Buckets {
		fmt.Println("Buckets:", bucket.Name)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP)
	go func() {
		<-ch
		os.Exit(0)
	}()

	for {
		xerror.Exit(db.Update(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchSize = 10

			it := txn.NewIterator(opts)
			defer it.Close()

			//for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				k := item.Key()

				if !bytes.HasPrefix(k, []byte(prefix)) {
					continue
				}

				var val []byte
				if err := item.Value(func(v []byte) error {
					val = v
					return nil
				}); err != nil {
					return err
				}

				var wf WatcherFile
				xerror.Panic(jsoniter.Unmarshal(val, &wf))

				fmt.Printf("%#v\n", wf)
				if !wf.Changed || wf.Synced {
					continue
				}

				fmt.Printf("key=%s, value=%s\n", k, val)
				xerror.Panic(kk.PutObjectFromFile("watcher_file"+wf.Path, wf.Path))
				wf.Synced = true
				wf.Changed = false
				wf.Hash = getHash(wf.Path)
				xerror.Panic(txn.Set(k, getBytes(&wf)))
				//kk.GetObjectMeta()
			}

			return nil
		}))

		log.Println("checking......")
		time.Sleep(time.Minute)
	}
}

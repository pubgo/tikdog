package main

import (
	"encoding/hex"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	badger "github.com/dgraph-io/badger/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/pubgo/tikdog/tikdog_cron"
	"github.com/pubgo/tikdog/tikdog_watcher"
	"github.com/pubgo/xerror"
	"github.com/twmb/murmur3"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

type SyncFile struct {
	Name    string
	Path    string
	Hash    string
	Changed bool
	Synced  bool
	Size    int64
	Mode    os.FileMode
	ModTime int64
	IsDir   bool
}

func getBytes(data interface{}) []byte {
	dt, _ := jsoniter.Marshal(data)
	return dt
}

func getHash(path string) (hash string) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("failed to open %s\n", path)
		return ""
	}
	defer file.Close()

	var h_ob = murmur3.New64()
	if _, err = io.Copy(h_ob, file); err == nil {
		hash := h_ob.Sum(nil)
		hashvalue := hex.EncodeToString(hash)
		return hashvalue
	} else {
		fmt.Printf("failed to sum %s\n", err)
		return "something wrong when use sha256 interface..."
	}
}

// 本地文件加载
// 本地存储中，如果已经同步了，那么就不用同步了
//

var prefix = "sync_files"
var ext = "drawio"

func syncDir(dir string, kk *oss.Bucket, db *badger.DB, ext string) {
	var sfs []SyncFile
	xerror.Panic(filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		if info.Name() == ".DS_Store" {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ext) {
			return nil
		}

		key := []byte(filepath.Join(prefix, path))

		return xerror.Wrap(db.View(func(txn *badger.Txn) error {
			itm, err := txn.Get(key)
			if err == badger.ErrKeyNotFound {
				sfs = append(sfs, SyncFile{
					Name:    info.Name(),
					Size:    info.Size(),
					Mode:    info.Mode(),
					ModTime: info.ModTime().Unix(),
					IsDir:   info.IsDir(),
					Synced:  false,
					Changed: true,
					Path:    path,
					Hash:    getHash(path),
				})
				return nil
			}
			xerror.Panic(err)

			xerror.Panic(itm.Value(func(_val []byte) error {
				var sf SyncFile
				xerror.Panic(jsoniter.Unmarshal(_val, &sf))
				if sf.ModTime != info.ModTime().Unix() {
					sf.Name = info.Name()
					sf.Size = info.Size()
					sf.Mode = info.Mode()
					sf.ModTime = info.ModTime().Unix()
					sf.IsDir = info.IsDir()
					sf.Changed = true
					sf.Hash = getHash(path)
				}

				sfs = append(sfs, sf)
				return nil
			}))
			return nil
		}))
	}))

	for i := range sfs {
		sf := sfs[i]
		key := filepath.Join(prefix, sf.Path)

		if !sf.Synced {
			fmt.Println("sync:", key, sf.Path)
			xerror.Panic(kk.PutObjectFromFile(key, sf.Path))
			sf.Synced = true
		}

		if sf.Changed {
			xerror.Panic(db.Update(func(txn *badger.Txn) error {
				sf.Changed = false
				fmt.Println("store:", key, sf.Path)
				return xerror.Wrap(txn.Set([]byte(key), getBytes(sf)))
			}))
		}
	}
}

func main() {
	tikdog_watcher.Start()
	xerror.Exit(tikdog_cron.Start())
	defer tikdog_cron.Stop()

	client, err := oss.New(
		os.Getenv("oss_endpoint"),
		os.Getenv("oss_ak"),
		os.Getenv("oss_sk"),
	)
	xerror.Panic(err)

	kk := xerror.PanicErr(client.Bucket("kooksee")).(*oss.Bucket)

	db, err := badger.Open(badger.DefaultOptions("demo"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	xerror.Exit(tikdog_cron.Add("Documents", "0 0/1 * * * *", func(event tikdog_cron.Event) error {
		syncDir(os.ExpandEnv("${HOME}/Documents"), kk, db, ext)
		return event.Err()
	}))

	xerror.Exit(tikdog_cron.Add("Downloads", "0 0/1 * * * *", func(event tikdog_cron.Event) error {
		syncDir(os.ExpandEnv("${HOME}/Downloads"), kk, db, "")
		return event.Err()
	}))

	xerror.Exit(tikdog_cron.Add("git/docs", "0 0/1 * * * *", func(event tikdog_cron.Event) error {
		syncDir(os.ExpandEnv("${HOME}/git/docs"), kk, db, "")
		return event.Err()
	}))

	//for _, k := range xerror.PanicErr(kk.ListObjectsV2(oss.Prefix(prefix))).(oss.ListObjectsResultV2).Objects {
	//	fmt.Printf("%#v\n", k)
	//}

	lsRes, err := client.ListBuckets()
	xerror.Panic(err)

	for _, bucket := range lsRes.Buckets {
		fmt.Println("Buckets:", bucket.Name)
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP)
	<-ch

	//for {
	//	xerror.Exit(db.Update(func(txn *badger.Txn) error {
	//		opts := badger.DefaultIteratorOptions
	//		opts.PrefetchSize = 10
	//
	//		it := txn.NewIterator(opts)
	//		defer it.Close()
	//
	//		//for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
	//		for it.Rewind(); it.Valid(); it.Next() {
	//			item := it.Item()
	//			k := item.Key()
	//
	//			if !bytes.HasPrefix(k, []byte(prefix)) {
	//				continue
	//			}
	//
	//			var val []byte
	//			if err := item.Value(func(v []byte) error {
	//				val = v
	//				return nil
	//			}); err != nil {
	//				return err
	//			}
	//
	//			var wf SyncFile
	//			xerror.Panic(jsoniter.Unmarshal(val, &wf))
	//
	//			fmt.Printf("%#v\n", wf)
	//			if !wf.Changed || wf.Synced {
	//				continue
	//			}
	//
	//			fmt.Printf("key=%s, value=%s\n", k, val)
	//			xerror.Panic(kk.PutObjectFromFile("watcher_file"+wf.Path, wf.Path))
	//			wf.Synced = true
	//			wf.Changed = false
	//			wf.Hash = getHash(wf.Path)
	//			xerror.Panic(txn.Set(k, getBytes(&wf)))
	//			//kk.GetObjectMeta()
	//		}
	//
	//		return nil
	//	}))
	//
	//	log.Println("checking......")
	//	time.Sleep(time.Minute)
	//}
}

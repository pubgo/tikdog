package tikdog_sync

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"go.uber.org/atomic"
	"hash/crc64"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	badger "github.com/dgraph-io/badger/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/pubgo/tikdog/internal/config"
	"github.com/pubgo/tikdog/tikdog_cron"
	"github.com/pubgo/tikdog/tikdog_watcher"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xprocess"
	"github.com/spf13/cobra"
	"github.com/twmb/murmur3"
)

type SyncFile struct {
	Crc64ecma uint64
	Name      string
	Path      string
	Changed   bool
	Synced    bool
	Size      int64
	Mode      os.FileMode
	ModTime   int64
	IsDir     bool
}

func handleKey(key string) string {
	return strings.ReplaceAll(key, " ", "-")
}

func getBytes(data interface{}) []byte {
	dt, _ := jsoniter.Marshal(data)
	return dt
}

func Hash(data []byte) (hash string) {
	var h = murmur3.New64()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func getHash(path string) (hash uint64) {
	var n = time.Now()
	defer func() {
		fmt.Println(time.Since(n), path)
	}()

	dt, err := ioutil.ReadFile(path)
	xerror.Panic(err)

	c := crc64.New(crc64.MakeTable(crc64.ECMA))
	xerror.PanicErr(c.Write(dt))
	return c.Sum64()
}

// 本地文件加载
// 本地存储中，如果已经同步了，那么就不用同步了
//

var prefix = "sync_files"
var ext = "drawio"

func syncDir(dir string, kk *oss.Bucket, db *badger.DB, ext string, c *atomic.Uint32) {
	fmt.Println("checking", dir)

	var sfs = make(chan SyncFile, 1)

	xprocess.GoLoop(func(ctx context.Context) {
		sf, ok := <-sfs
		if !ok {
			return
		}

		c.Inc()

		key := filepath.Join(prefix, sf.Path)

		if !sf.Synced {
			head, err := kk.GetObjectMeta(key)
			xerror.Panic(err)

			ccc, err := strconv.ParseUint(head.Get("X-Oss-Hash-Crc64ecma"), 10, 64)
			xerror.Panic(err)
			if ccc != sf.Crc64ecma {
				fmt.Println("sync:", key, sf.Path)
				xerror.Exit(kk.PutObjectFromFile(key, sf.Path))
			}
			sf.Changed = true
			sf.Synced = true
		}

		if sf.Changed {
			xerror.Exit(db.Update(func(txn *badger.Txn) error {
				sf.Changed = false
				fmt.Println("store:", key, sf.Path)
				return xerror.Wrap(txn.Set([]byte(Hash([]byte(key))), getBytes(sf)))
			}))
		}
	})

	xerror.Exit(filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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
			itm, err := txn.Get([]byte(Hash(key)))
			if err == badger.ErrKeyNotFound {
				fmt.Println("ErrKeyNotFound:", string(key))

				sfs <- SyncFile{
					Name:      info.Name(),
					Size:      info.Size(),
					Mode:      info.Mode(),
					ModTime:   info.ModTime().Unix(),
					IsDir:     info.IsDir(),
					Synced:    false,
					Changed:   true,
					Path:      path,
					Crc64ecma: getHash(path),
				}
				return nil
			}

			xerror.Panic(err)

			xerror.Panic(itm.Value(func(_val []byte) error {
				var sf SyncFile
				xerror.Panic(jsoniter.Unmarshal(_val, &sf))
				if sf.ModTime == info.ModTime().Unix() {
					return nil
				}

				sf.Name = info.Name()
				sf.Size = info.Size()
				sf.Mode = info.Mode()
				sf.ModTime = info.ModTime().Unix()
				sf.IsDir = info.IsDir()
				sf.Changed = true

				hash := getHash(path)
				if sf.Crc64ecma != hash {
					sf.Synced = false
					sf.Crc64ecma = hash
				}

				sfs <- sf
				return nil
			}))
			return nil
		}))
	}))
}

func GetCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "sync", Short: "sync from local to remote"}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		defer xerror.RespDebug()

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

		opts := badger.DefaultOptions(filepath.Join(config.Home, "db"))
		db, err := badger.Open(opts)
		xerror.Panic(err)
		defer db.Close()

		var nw = NewWaiter()
		xerror.Exit(tikdog_cron.Add("Documents", "0/5 * * * * *", func(event tikdog_cron.Event) error {
			key := os.ExpandEnv("${HOME}/Documents")

			if nw.Skip(key) {
				return nil
			}

			var c = atomic.NewUint32(0)
			syncDir(key, kk, db, ext, c)
			nw.Report(key, c)
			return event.Err()
		}))

		xerror.Exit(tikdog_cron.Add("Downloads", "0/5 * * * * *", func(event tikdog_cron.Event) error {
			key := os.ExpandEnv("${HOME}/Downloads")

			if nw.Skip(key) {
				return nil
			}

			var c = atomic.NewUint32(0)
			syncDir(key, kk, db, "", c)
			nw.Report(key, c)
			return event.Err()
		}))

		xerror.Exit(tikdog_cron.Add("git/docs", "0/5 * * * * *", func(event tikdog_cron.Event) error {
			key := os.ExpandEnv("${HOME}/git/docs")

			if nw.Skip(key) {
				return nil
			}

			var c = atomic.NewUint32(0)
			syncDir(key, kk, db, "", c)
			nw.Report(key, c)
			return event.Err()
		}))

		//for _, k := range xerror.PanicErr(kk.ListObjectsV2(oss.Prefix(prefix))).(oss.ListObjectsResultV2).Objects {
		//	fmt.Printf("%#v\n", k)
		//}

		//lsRes, err := client.ListBuckets()
		//xerror.Panic(err)

		//for _, bucket := range lsRes.Buckets {
		//	fmt.Println("Buckets:", bucket.Name)
		//}

		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP)
		<-ch
	}

	cmd.AddCommand(GetDbCmd())
	return cmd
}

func GetDbCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "db"}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		var prefix string
		if len(args) > 0 {
			prefix = args[0]
		}

		//var code = "true"
		//if len(args) > 1 {
		//	code = args[1]
		//}

		//program, err := expr.Compile(code, expr.Env(&SyncFile{}))
		//xerror.Panic(err)

		dbPath := filepath.Join(config.Home, "db")
		opts := badger.DefaultOptions(dbPath)
		opts.WithLoggingLevel(badger.DEBUG)

		db, err := badger.Open(opts)
		xerror.Panic(err)
		defer db.Close()

		xerror.Exit(db.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchSize = 10

			it := txn.NewIterator(opts)
			defer it.Close()

			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()

				if !bytes.HasPrefix(item.Key(), []byte(prefix)) {
					continue
				}

				xerror.Panic(item.Value(func(v []byte) error {
					var sf SyncFile
					xerror.Panic(jsoniter.Unmarshal(v, &sf))
					//output, err := expr.Run(program, &sf)
					//xerror.Panic(err)
					//if output.(bool) {
					fmt.Println(string(item.Key()), string(v))
					//}

					return nil
				}))
			}

			return nil
		}))
	}
	return cmd
}

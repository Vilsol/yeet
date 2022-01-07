package tests

import (
	"context"
	"github.com/MarvinJWendt/testza"
	"github.com/Vilsol/yeet/cache"
	"github.com/Vilsol/yeet/flat/yeet"
	"github.com/Vilsol/yeet/source"
	"github.com/Vilsol/yeet/utils"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/rs/zerolog"
	cuckoo "github.com/seiflotfy/cuckoofilter"
	"github.com/spf13/viper"
	"io"
	"os"
	"testing"
)

const host = "localhost:8080"

func init() {
	viper.Set("paths", []string{"."})
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func BuildSample() []byte {
	dir, _ := os.ReadDir("../docs/")
	filterSize := utils.EstimateCuckooFilter(int64(len(dir)))
	filter := cuckoo.NewFilter(filterSize)

	for _, entry := range dir {
		filter.InsertUnique([]byte("/" + entry.Name()))
	}

	builder := flatbuffers.NewBuilder(1024)
	bucket := builder.CreateString("yeet")
	endpoint := builder.CreateString("http://localhost:9000")
	accessKey := builder.CreateString("minio")
	secretKey := builder.CreateString("minio123")
	region := builder.CreateString("us-west-002")
	encodedFilter := builder.CreateByteString(filter.Encode())

	yeet.S3Start(builder)
	yeet.S3AddBucket(builder, bucket)
	yeet.S3AddEndpoint(builder, endpoint)
	yeet.S3AddKey(builder, accessKey)
	yeet.S3AddSecret(builder, secretKey)
	yeet.S3AddRegion(builder, region)
	yeet.S3AddFilter(builder, encodedFilter)
	s3 := yeet.S3End(builder)
	builder.Finish(s3)

	return builder.FinishedBytes()
}

func GetRedis() (*source.S3Redis, error) {
	s, err := source.NewS3Redis("tcp", "localhost:6379", "", "", 0)
	if err != nil {
		return nil, err
	}

	err = s.RedisClient.Set(context.Background(), "yeet:"+host, BuildSample(), 0).Err()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func TestS3Redis(t *testing.T) {
	s, err := GetRedis()
	testza.AssertNil(t, err)

	c, err := cache.NewHashMapCache(s, true)
	testza.AssertNil(t, err)

	dir, _ := os.ReadDir("../docs/")
	for _, entry := range dir {
		file, _ := os.ReadFile("../docs/" + entry.Name())

		fileType, reader, size := c.Get([]byte("/"+entry.Name()), []byte(host))

		testza.AssertEqual(t, "text/markdown; charset=utf-8", fileType)
		testza.AssertNotNil(t, reader)
		testza.AssertEqual(t, len(file), size)

		responseBody, err := io.ReadAll(reader)

		testza.AssertNil(t, err)
		testza.AssertNil(t, reader.(io.Closer).Close())
		testza.AssertEqual(t, file, responseBody)
	}

	fileType, reader, size := c.Get([]byte("/does-not-exist.md"), []byte(host))
	testza.AssertEqual(t, "", fileType)
	testza.AssertNil(t, reader)
	testza.AssertEqual(t, 0, size)

	fileType, reader, size = c.Get([]byte("/yeet.md"), []byte("does-not-exist"))
	testza.AssertEqual(t, "", fileType)
	testza.AssertNil(t, reader)
	testza.AssertEqual(t, 0, size)
}

func BenchmarkS3Redis(b *testing.B) {
	s, err := GetRedis()
	testza.AssertNil(b, err)

	c, err := cache.NewHashMapCache(s, true)
	testza.AssertNil(b, err)

	hostBytes := []byte(host)
	path := []byte("/yeet.md")

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, reader, _ := c.Get(path, hostBytes)
			io.ReadAll(reader)
			if r, ok := reader.(io.Closer); ok {
				r.Close()
			}
		}
	})
}

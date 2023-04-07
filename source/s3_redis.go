package source

import (
	"context"
	"github.com/Vilsol/yeet/flat/yeet"
	"github.com/Vilsol/yeet/utils"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/seiflotfy/cuckoofilter"
	"time"
)

const redisPrefix = "yeet:"

var _ Source = (*S3Redis)(nil)

type S3Redis struct {
	RedisClient     *redis.Client
	CredentialCache *cache.Cache
}

type S3Wrapper struct {
	S3
	Filter *cuckoo.Filter
}

func NewS3Redis(network string, address string, username string, password string, db int) (*S3Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Network:  network,
		Addr:     address,
		Username: username,
		Password: password,
		DB:       db,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, errors.Wrap(err, "failed to connect to redis")
	}

	// TODO Configurable timeout and cleanup interval
	return &S3Redis{
		RedisClient:     rdb,
		CredentialCache: cache.New(time.Second*10, time.Minute),
	}, nil
}

func (s S3Redis) Get(path string, host []byte) (*utils.StreamHijacker, bool) {
	var s3Wrapper *S3Wrapper

	if host == nil {
		return nil, true
	}

	if instance, ok := s.CredentialCache.Get(utils.ByteSliceToString(host)); ok {
		s3Wrapper = instance.(*S3Wrapper)
	} else {
		get := s.RedisClient.Get(context.TODO(), redisPrefix+utils.ByteSliceToString(host))
		if get.Err() != nil {
			if errors.Is(get.Err(), redis.Nil) {
				log.Warn().Str("host", utils.ByteSliceToString(host)).Msg("no credentials found")
				return nil, true
			}

			log.Error().Err(get.Err()).Msg("failed to get credentials")
			return nil, true
		}

		s3Flat := yeet.GetRootAsS3(utils.UnsafeGetBytes(get.Val()), 0)

		s3Instance, err := NewS3(
			utils.ByteSliceToString(s3Flat.Bucket()),
			utils.ByteSliceToString(s3Flat.Key()),
			utils.ByteSliceToString(s3Flat.Secret()),
			utils.ByteSliceToString(s3Flat.Endpoint()),
			utils.ByteSliceToString(s3Flat.Region()),
		)

		if err != nil {
			log.Err(err).Msg("failed to create new S3 session")
			return nil, true
		}

		cf, err := cuckoo.Decode(s3Flat.Filter())
		if err != nil {
			log.Err(err).Msg("failed to decode filter")
			return nil, true
		}

		s3Wrapper = &S3Wrapper{
			S3:     *s3Instance,
			Filter: cf,
		}

		s.CredentialCache.Set(utils.ByteSliceToString(host), s3Wrapper, cache.DefaultExpiration)
	}

	if s3Wrapper.Filter.Lookup(utils.UnsafeGetBytes(path)) {
		return GetS3(s3Wrapper.S3Client, s3Wrapper.Bucket, path)
	}

	return nil, false
}

func (s S3Redis) IndexPath(_ string, _ IndexFunc) (int64, int64, error) {
	// Indexing not supported for Redis-backed S3
	return 0, 0, nil
}

func (s S3Redis) Watch() (<-chan WatchEvent, error) {
	return nil, errors.New("s3-redis does not support watching")
}

package configs

import "testing"

func TestLoad(t *testing.T) {
	wantPort := "8080"

	t.Setenv("PORT", wantPort)
	t.Setenv("MAX_UPLOAD_SIZE_BYTES", "8080")
	t.Setenv("REDIS_URL", "redis.com")
	t.Setenv("STREAM_NAME", "stream")
	t.Setenv("WORKER_GROUP_NAME", "worker")
	t.Setenv("SEMAPHORE_SIZE", "10")

	cfg := Load()

	port := cfg.Port
	if port == "" {
		t.Fatal("port is missing")
	}

	if port != wantPort {
		t.Fatal("invalid port value")
	}
}
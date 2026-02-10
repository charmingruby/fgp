# Config Validation with Validated

## Problem
Load configuration from multiple sources and report all invalid fields at once instead of failing fast on the first error.

## Functional Design
```go
type AppConfig struct {
	Host string
	Port int
}

func parse(cfg rawConfig) validated.Validated[error, AppConfig] {
	host := func() validated.Validated[error, string] {
		if cfg.Host == "" {
			return validated.Invalid[error, string](errors.New("host missing"))
		}
		return validated.Valid[error](cfg.Host)
	}()

	port := func() validated.Validated[error, int] {
		if cfg.Port <= 0 {
			return validated.Invalid[error, int](errors.New("port invalid"))
		}
		return validated.Valid[error](cfg.Port)
	}()

	return validated.Map(
		validated.Zip(host, port),
		func(pair result.Tuple2[string, int]) AppConfig {
			return AppConfig{Host: pair.First, Port: pair.Second}
		},
	)
}

func load(raw rawConfig) (AppConfig, error) {
	val := parse(raw)
	return validated.ToResult(val).Unwrap()
}
```

## Imperative Comparison
Conventional validation chains often accumulate errors manually using slices and conditionals. `Validated` provides structural accumulation: `Zip` merges failures from each field, `Sequence` handles slices of nested configs, and `ToResult` bridges to existing `(T, error)` consumers.

## Performance Notes
- Errors are stored inline in a slice; no interfaces or reflection are used.
- `Invalid` copies the error slice once so subsequent combinators can share without mutation.
- `Validated.ToResult` uses `errors.Join` to return a single error compatible with standard logging pipelines.

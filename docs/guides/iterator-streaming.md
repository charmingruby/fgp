# Streaming Ingestion with Lazy Iterators

## Problem
Ingest an unbounded stream of events from Kafka while applying filters, projections, and batching without creating goroutines per stage.

## Functional Design
```go
func stream(source kafka.Reader) seq.Iterator[Event] {
	return seq.Iterator[Event]{
		next: func() (Event, bool) {
			msg, err := source.Fetch(context.Background())
			if err != nil {
				return Event{}, false
			}
			return decode(msg.Value), true
		},
	}
}

pipeline := seq.TakeWhile(
	seq.MapIter(
		seq.FilterIter(stream(reader), func(e Event) bool { return e.Kind == "purchase" }),
		func(e Event) Processed { return project(e) },
	),
	func(p Processed) bool { return !p.Stale },
)

for chunk := range seq.Chunk(seq.ToSlice(seq.Take(pipeline, 500)), 50) {
	flush(chunk)
}
```

## Imperative Comparison
Imperative ingestion often spawns goroutines per stage or channels between filters. Pull-based iterators keep control in the consumer, so there are no queues to size or goroutines to leak. Composition stays ergonomic by using `MapIter`, `FilterIter`, `TakeWhile`, and `Chunk`.

## Performance Notes
- All iterator helpers are closure-based; no goroutines or extra channels are created by the combinators themselves.
- Back-pressure is automatic: `Next` drives the stream, so the consumer decides throughput.
- `Chunk` copies data to avoid aliasing, guaranteeing immutability for downstream processing.

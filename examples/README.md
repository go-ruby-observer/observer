# Ruby examples

Pure-Ruby examples of `observer` — the `Observable` publish/subscribe mixin this
library backs. They run under [go-embedded-ruby](https://github.com/go-embedded-ruby/ruby)
(rbgo) via the `require "observer"` binding.

```sh
rbgo examples/observer_usage.rb
```

| File | Shows |
| --- | --- |
| [`observer_usage.rb`](observer_usage.rb) | Including `Observable`, `add_observer`, `count_observers`, the `changed`/`changed?`/`notify_observers` lifecycle (dispatch in insertion order then reset), `delete_observers`, and notify as a no-op when not changed. |

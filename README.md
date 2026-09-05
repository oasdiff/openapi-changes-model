# The OpenAPI Changes Model

OpenAPI gave API definitions a formal model. This repository does the same for **API changes**: a machine-readable classification of every possible change to an OpenAPI contract, and whether each one breaks existing API consumers.

The model is one file, [`openapi-changes-model.yaml`](openapi-changes-model.yaml). It contains:

- **The vocabulary**: the seven actions a document edit can perform (add, remove, set, unset, increase, decrease, change), the two wire directions, the six effects a change can have on the set of valid payloads, and the guards, document states that qualify a verdict.
- **The severity law**: one rule that derives every verdict. Guards apply first; then narrowing breaks request consumers, widening breaks response consumers, an incomparable change breaks both, and an unknown one is a warning. When a change cannot be proven safe, it is breaking.
- **681 named changes**, each with its direction, effect, guards, derived severity, human-readable message, and its **claims**: the exact document locations and actions it covers, as `location:action` patterns over the OpenAPI object model.
- **The full edit space**: 15,255 possible edits, enumerated mechanically from the OpenAPI specification's object model. Every edit is covered by named changes, waived with a written reason, or classified as non-contract (unable to affect which payloads are valid). None are undecided.

See [SPEC.md](SPEC.md) for the concepts in prose.

## Status

**Draft (0.1.0).** The model is exported from [oasdiff](https://github.com/oasdiff/oasdiff), its reference implementation, at every oasdiff release; the generator in [`generator/`](generator/) pins the exact version. The direction of authority is deliberate for now: verdicts are debated here, decided changes land in the reference implementation, and the next export carries them. As the model stabilizes, the intent is to reverse this: the model becomes the source of truth and implementations, oasdiff included, derive from it.

## Contributing

Disagreement with a verdict is the most valuable contribution. If you believe a change is classified wrongly, wrongly waived, or missing:

- **Open an issue** naming the change id (or the edit's location and action) and the payload that demonstrates your case: a request or response that was valid before and breaks after, or the reason none can exist.
- **Pull requests** are welcome against `SPEC.md` and the documentation. The model file itself is generated, so classification changes are agreed in an issue first and then land via the reference implementation.

The severity law is the part most worth challenging: it is short, explicit, and every one of the 681 verdicts follows from it. An argument that changes the law changes everything downstream, on purpose.

## Reading the model

Each named change carries `claims`, the edits it covers:

```yaml
- id: request-property-max-length-decreased
  level: error
  direction: request
  effect: narrows
  claims:
    - paths.*.*.requestBody.content.*.schema.maxLength:decrease
  message: the %s request property's maxLength was decreased to %s
```

The `coverage` section is the exhaustiveness proof: every edit's disposition, including the written reason for every exclusion. The reference implementation's build fails if an edit is left undecided or a reason goes stale.

## License

[Apache-2.0](LICENSE), the same license as the OpenAPI Specification.

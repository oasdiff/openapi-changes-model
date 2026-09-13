# The OpenAPI Changes Model: Concepts

*Version 0.1.0-draft. The normative artifact is [`openapi-changes-model.yaml`](openapi-changes-model.yaml); this document explains its concepts.*

## The contract

An OpenAPI document declares a contract: which requests are valid and which responses a client can receive. A change is judged against that contract, not against how tolerantly any particular server behaves. A change is **breaking** when a consumer that conformed to the old contract can stop conforming, or fail, under the new one. Consumers include runtime clients, generated SDKs, gateways, validators, and AI agents built from the contract.

## Edits

An **edit** is one syntactic change at one location of the document: a `(location, action)` pair. Locations are paths through the OpenAPI object model (`paths.*.*.requestBody.content.*.schema.maxLength`); the `*` segments stand in for names the API author chooses (a path, a method, a media type). Actions are:

| action | meaning |
|---|---|
| `add` / `remove` | a member joins or leaves a collection |
| `set` / `unset` | a field appears where it was absent, or disappears |
| `increase` / `decrease` | an ordered field's value grows or shrinks |
| `change` | a field's value is replaced by an incomparable value |

The edit space is enumerated mechanically from the object model, so it is complete by construction: 15,255 edits as of this version, growing only when the OpenAPI specification itself grows.

The location syntax is this model's own, deliberately. JSONPath (RFC 9535), which the OpenAPI Overlay Specification uses for its targets, expresses most of these paths directly, but not all of them: locations here address the object model rather than a document instance (the `schema` segment covers the schema and its sub-schemas at any depth), and the `x-*` form matches key names by prefix, which JSONPath expresses only through filter expressions. Aligning with JSONPath where the semantics permit is an open question; see the issue tracker.

## Changes

A **change** is a named, human-meaningful classification of one or more edits, identified by an id (`request-property-max-length-decreased`). Each change declares:

- **claims**: the edits it covers, as `location:action[,action...]` patterns;
- **direction**: whether it concerns what clients send (`request`), receive (`response`), or neither;
- **effect**: what the change does to the set of valid payloads: `narrows`, `widens`, `incomparable`, `none`, `unknown` (the specification cannot decide), or `violation` (a lifecycle contract is broken);
- **guards**: document states that qualify the verdict (a `readOnly` property never appears in requests; an honored sunset sanctions a removal);
- **level**: the derived severity;
- **area** and **kind**: where the change sits in the OpenAPI object model and which aspect of the contract it touches, for querying and auditing the catalog.

## The severity law

Every level is derived, never assigned. Guards apply first, each nullifying or requalifying the effect on the side it speaks about. Then:

| effect | direction | level |
|---|---|---|
| narrows | request | error |
| narrows | response | info |
| widens | request | info |
| widens | response | error |
| incomparable | any | error |
| violation | any | error |
| unknown | any | warning |
| none | any | info |

The asymmetry is deliberate: reporting a safe change as breaking costs a reviewer one look, while reporting a breaking change as safe ships it to production. A change is declared safe only when it is provably safe for every consumer that conformed to the old contract; any gap in that proof resolves to breaking. Where the specification itself lacks the information to decide, the verdict is a warning that says what is missing, never a guess.

## Transitions

Some edits arrive together as one semantic change. Wrapping a schema in `oneOf: [{type: "null"}, X]` to make it nullable is a single decision, but the raw diff shows several edits: the type changed, an enum moved, a `oneOf` appeared. A **transition** names such a shape. At a recognized shape, raw findings of the transition's *claimed kinds* are echoes of the one change and are suppressed; findings of other kinds are independent changes and still report. The transition itself is reported by its listed changes, so nothing is silently dropped: the finding moves from the echoes to the recognition.

## Coverage dispositions

Every edit in the space has exactly one disposition:

- **covered**: one or more named changes claim it;
- **waived**: no change covers it, and a written reason says why (`open`: a missing check with its reason and a suggested id; `resolved-at-usage`: component definitions are compared where they are referenced; `covered-as`: the same document edit is reported under another action);
- **non-contract**: the edit cannot affect which payloads are valid (descriptions, examples, extensions).

No edit is undecided. The reference implementation enforces this in its build: an unclassified edit, a stale waiver, or a severity that deviates from the law is a failing test, not a review comment.

## The boundary

The contract is the boundary of the model. A change in behavior the document does not describe, the meaning of a field, a rate limit, an authorization rule outside the spec, is invisible to any specification-based classification. That is a reason to keep the document faithful to the API, not a gap in the model's own promise.

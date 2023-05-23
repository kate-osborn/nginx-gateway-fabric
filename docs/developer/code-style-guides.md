# Code Style Guides

## Git

* Keep a clean, concise and meaningful git commit history on your branch, rebasing locally and squashing before
  submitting a PR
* Follow the guidelines of writing a good commit message as described [here](pull-request.md#commit-message-format)

## Docker

Follow industry “best practices”: https://docs.docker.com/engine/userguide/eng-image/dockerfile_best-practices/

## Go

Before diving into project-specific guidelines, please familiarize yourself with the following vetted best practices:
- [Effective Go](https://go.dev/doc/effective_go)
- [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [100 Go Mistakes](https://github.com/teivah/100-go-mistakes)

Once you have a good grasp of these general best practices, you can then explore the project-specific guidelines for
your particular project. These guidelines will often build upon the foundation set by the general best practices and
provide additional recommendations tailored to the project's specific requirements and coding style.

### Best Practices

#### Use the empty struct `struct{}` for sentinel values

Empty structs as sentinels unambiguously signal an explicit lack of information. For example, use empty struct when for
sets and for signaling via channels that don't require a message.

DO:

```go
set := make(map[string]struct{}) // empty struct is empty struct // nearly value-less

signaller := make(chan struct{}, 0)
signaller <- struct{}{} // no information but signal on // delivery
```

DO NOT:

```go
set := make(map[string]bool) // is true/false meaningful? // is this a set?

signaller := make(chan bool, 0)
signaller <- false // is this a signal? is this an error?
```

#### Consistent Line Breaks

When breaking up a long function definition, call, or struct initialization, choose to break after each parameter,
argument, or field.

DO:

```go
func longFunctionDefinition(
    paramX int,
    paramY string,
    paramZ bool,
) (string, error){}

// and 

s := myStruct{
    field1: 1,
    field2: 2,
    field3: 3,
}
```

DO NOT:

```go
func longFunctionDefinition(paramX int, paramY string,
    paramZ bool,
) (string, error){}

// or 

func longFunctionDefinition(
    paramX int, paramY string,
    paramZ bool,
) (string, error){}

// or 
s := myStruct{field1: 1, field2: 2,
    field3: 3}

```

When constructing structs pass members during initialization.

Example:

```go
cfg := foo.Config{
    Site: "example.com",
    Out: os.Stdout,
    Dest: c.KeyPair{
        Key: "style",
        Value: "well formatted",
    },
}
```

#### Do not copy sync entities

`sync.Mutex` and `sync.Cond` MUST NOT be copied. By extension, structures holding an instance MUST NOT be copied. By
extension, structures which embed instances MUST NOT be copied.

DO NOT embed sync entities. Pointers to `Mutex` and `Cond` are required for storage.

#### Construct slices with known capacity

Whenever possible bounded slices should be constructed with a length of zero size, but known capacity.

```go
s := make([]string, 0, 32)
```

Growing a slice is an expensive deep copy operation. When the bounds of a slice can be calculated, pre-allocating the
storage allows for append to assign a value without allocating new memory.

#### Accept interfaces and return structs

Structures are expected to be return values from functions. If they satisfy an interface, any struct may be used in
place of that interface. Interfaces will cause escape analysis and likely heap allocation when returned from a
function - concrete instances (as copies) may stay as stack memory.

Returning interfaces from functions will hide the underlying structure type. This can lead to unintended growth of the
interface type when methods are needed but unavailable, or an API change to return the structure later.

Accepting interfaces as arguments ensures forward compatibility as API responsibilities grow. Structures as arguments
will require additional functions or breaking API changes. Whereas interfaces inject behavior and may be replaced or
modified without changing the signature.

#### Use contexts in a viral fashion

Functions that accept a `context.Contex`t should pass the context or derive a subsidiary context to functions it calls.
When designing libraries, subordinate functions (especially those asynchronous or expensive in nature) should accept
`context.Context`.

Public APIs SHOULD be built from inception to be context aware. ALL asynchronous public APIs MUST be built from 
inception to be context aware.

#### Do not use templates to replace interface types

Use _templates_ to substitute for concrete types, use _interfaces_ to substitute for abstract behaviors. 

#### Do NOT use booleans as function parameters 

A boolean can only express an on/off condition and will require a breaking change or new arguments in the future. 
Instead, pack functionality; for example, using integers instead of bools (maps and structs are also acceptable).

DO NOT:

```go
func(bool userOn, bool groupOn, bool globalOn)
```

DO:

```go
func(uint permissions) // permissions := USER | GROUP | GLOBAL // if permissions & USER { // USER is set }
```

#### Use dependency injection to separate concerns

Functions, structures, interfaces and modules are designed as clients that are provided services by injection. Clients
use services without the need for creation.

Creating dependencies couples ownership and lifetime while making tests difficult or impossible. Constructing
dependencies adds side-effects which complicates testing.

#### Required arguments should be provided via parameters and optional arguments provided functionally or with structs

DO NOT:

```go 
func(int required, int optional) { 
    if optional { 
        
    } 
}
```

DO:

```go
type Option func(o *Object)

func Optional(string optional) Option {
    return func(o *Object) {
        o.optional = optional 
    } 
} 

func(int required, ...Options) { 
    for o := range Options {
        o(self)
    } 
}
```


### Error Handling

#### Do not filter context when returning errors

Preserve error context by wrapping errors as the stack unwinds. Utilize native error wrapping with`fmt.Errorf` and 
the `%w` verb to wrap errors. Wrapped errors offer a transparent view to end users. For a practical example, 
refer to this runnable code snippet: [Go Playground Example](https://go.dev/play/p/f9EaJDB5JUO). When required, 
you can identify inner wrapped errors using native APIs such as `As`, `Is`, and `Unwrap`.


#### Only handle errors once

DO NOT log an error, then subsequently return that error. This creates the potential for multiple error reports and
conflicting information to users.

DO NOT:

```go
func badAtStuff(noData string) error {
    if len(noData) == 0 {
        fmt.Printf("Received no data")
    }
    
    return errors.New("received no data")
}
```

DO

```go
func badAtStuff(noData string) error {
    if len(noData) == 0 {
        return errors.New("received no data")
    }
    ...
}
```

A leader of the golang community once said:
> Lastly, I want to mention that you should only handle errors once. Handling an error means inspecting the
> error value, and making a decision. If you make less than one decision, you’re ignoring the error...But making
> more than one decision in response to a single error is also problematic. - Dave Cheney

#### Libraries should return errors for callers to handle

Asynchronous libraries should communicate via channels or callbacks. Only then should they log unhandled errors.

Example:

```go
func onError(err error) { 
    // got an asynchronous error 
}

func ReadAsync(r io.Reader, onError) { 
    err := r()
    if err != nil { 
        onError(err)
    } 
} 

go ReadAsync(reader, onError)

// OR errs := make(chan error)

func ReadAsync(r io.Reader, errs chan<- error) { 
    err := r()
    if err != nil { 
      // put error on errs channel, but don't block forever.
    }
}
```

#### Callers should handle errors and pass them up the stack for notification

Callers should handle errors that occur within the functions they call. This allows them to handle errors according to 
their specific requirements. However, if callers are unable to handle the error or need to provide additional context, 
they can add context to the error and pass it up the stack for notification. 

Example:

```go
func readFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

func processFile(filename string) error {
	data, err := readFile(filename)
	if err != nil {
		return fmt.Errorf("failed to process file: %w", err)
	}

	// Process the file data here

	return nil
}

func main() {
	filename := "example.txt"
	err := processFile(filename)
	if err != nil {
		fmt.Printf("Error processing file: %v\n", err) // caller handles the error
	}
}
```

### Logging

### Concurrency 

### Recommended / Situational 

These recommendations are generally related to performance and efficiency but will not be appropriate for all paradigms.

#### Use golang benchmark tests and pprof tools for profiling and identifying hot spots

The `-gcflags '-m'` can be used to analyze escape analysis and estimate logging costs.

#### Reduce the number of stored pointers. Structures should store instances whenever possible.

DO NOT use pointers to avoid copying. Pass by value. Ancillary benefit is reduction of nil checks. Fewer pointers helps
garbage collection and can indicate memory regions that can be skipped. It reduces de-referencing and bounds checking in
the VM. Keep as much on the stack as possible (see caveat: [accept interfaces](#accept-interfaces-and-return-structs) -
forward compatibility and flexibility concerns outweigh costs of heap allocation)

FAVOR:

```go
type Object struct{ 
    subobject SubObject 
} 

func New() Object { 
    return Object{ 
        subobject: SubObject{}, 
    }
}
```

DISFAVOR:

```go
type Object struct{ 
    subobject *SubObject 
} 

func New() *Object { 
    return &Object{ 
        subobject: &SubObject{}, 
    } 
}
```

#### Pass pointers down the stack not up 

Pointers can be passed down the stack without causing heap allocations. Passing pointers up the stack will cause heap
allocations.

```text
-> initialize struct A -> func_1(&A)
-> func_2(&A)
```

`A` can be passed as a pointer as it's passed down the stack.

Returning pointers can cause heap allocations and should be avoided.

DO NOT:

```go
func(s string) *string { 
    s := s + "more strings"
    return &s // this will move to heap 
}
```

#### Using interface types will cause unavoidable heap allocations

Frequently created, short-lived instances will cause heap and garbage collection pressure. Using a `sync.Pool` to store
and retrieve structures can improve performance.

Before adopting a `sync.Pool` analyze the frequency of creation and duration of lifetime. Objects created frequency for
short periods of time will benefit the most.

For example, channels that send and receive signals using interfaces; these are often only referenced for the duration of the
event and can be recycled once the signal is received and processed.

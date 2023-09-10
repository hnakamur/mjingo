package mjingo_test

import (
	"fmt"
	"log"
	"sync/atomic"

	"github.com/hnakamur/mjingo"
	"github.com/hnakamur/mjingo/option"
)

type Cycler struct {
	values []mjingo.Value
	idx    atomic.Uint64
}

var _ mjingo.Object = ((*Cycler)(nil))
var _ mjingo.Caller = ((*Cycler)(nil))

func (s *Cycler) Kind() mjingo.ObjectKind { return mjingo.ObjectKindPlain }

func (s *Cycler) Call(_state mjingo.State, _args []mjingo.Value) (mjingo.Value, error) {
	idx := int(s.idx.Add(1))
	return s.values[idx%len(s.values)], nil
}

func makeCycler(_state mjingo.State, args []mjingo.Value) (mjingo.Value, error) {
	return mjingo.ValueFromGoValue(&Cycler{
		values: args,
		idx:    atomic.Uint64{},
	}), nil
}

type Magic struct{}

var _ mjingo.Object = ((*Magic)(nil))
var _ mjingo.CallMethoder = ((*Magic)(nil))

func (s *Magic) Kind() mjingo.ObjectKind { return mjingo.ObjectKindPlain }

func (s *Magic) CallMethod(_state mjingo.State, name string, args []mjingo.Value) (mjingo.Value, error) {
	if name != "make_class" {
		return nil, mjingo.NewError(mjingo.UnknownMethod,
			fmt.Sprintf("object has no method named %s!!!", name))
	}

	if len(args) < 1 {
		return nil, mjingo.NewError(mjingo.MissingArgument, "")
	}
	if len(args) > 1 {
		return nil, mjingo.NewError(mjingo.TooManyArguments, "")
	}
	// single string argument
	tag, err := mjingo.ValueTryToGoValue[string](args[0])
	if err != nil {
		return nil, err
	}
	return mjingo.ValueFromGoValue(fmt.Sprintf("magic-%s", tag)), nil
}

type SimpleDynamicSec struct{}

var _ mjingo.Object = ((*SimpleDynamicSec)(nil))
var _ mjingo.SeqObject = ((*SimpleDynamicSec)(nil))

func (s *SimpleDynamicSec) Kind() mjingo.ObjectKind { return mjingo.ObjectKindSeq }

func (s *SimpleDynamicSec) GetItem(idx uint) option.Option[mjingo.Value] {
	if idx >= s.ItemCount() {
		return option.None[mjingo.Value]()
	}
	return option.Some(mjingo.ValueFromGoValue([]string{"a", "b", "c", "d"}[idx]))
}

func (s *SimpleDynamicSec) ItemCount() uint { return 4 }

const templateSource = `{%- with next_class = cycler(["odd", "even"]) %}
<ul class="{{ magic.make_class("ul") }}">
{%- for char in seq %}
  <li class={{ next_class() }}>{{ char }}</li>
{%- endfor %}
</ul>
{%- endwith %}`

func ExampleObject() {
	env := mjingo.NewEnvironment()
	const templateName = "test.html"
	env.AddFunction("cycler", mjingo.BoxedFuncFromFunc(makeCycler))
	env.AddGlobal("magic", mjingo.ValueFromGoValue(&Magic{}))
	env.AddGlobal("seq", mjingo.ValueFromGoValue(&SimpleDynamicSec{}))
	err := env.AddTemplate(templateName, templateSource)
	if err != nil {
		log.Fatal(err)
	}
	tpl, err := env.GetTemplate(templateName)
	if err != nil {
		log.Fatal(err)
	}
	context := mjingo.ValueFromGoValue(map[string]string{"title": "this is my page"})
	got, err := tpl.Render(context)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(got)
	// Output: <ul class="magic-ul">
	//   <li class=even>a</li>
	//   <li class=odd>b</li>
	//   <li class=even>c</li>
	//   <li class=odd>d</li>
	// </ul>
}

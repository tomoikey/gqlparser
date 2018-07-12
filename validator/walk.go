package validator

import (
	"context"
	"fmt"

	"github.com/vektah/gqlparser"
)

type Events struct {
	operationVisitor      []func(walker *Walker, operation *gqlparser.OperationDefinition)
	operationLeaveVisitor []func(walker *Walker, operation *gqlparser.OperationDefinition)
	field                 []func(walker *Walker, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field)
	fragment              []func(walker *Walker, parentDef *gqlparser.Definition, fragment *gqlparser.FragmentDefinition)
	inlineFragment        []func(walker *Walker, parentDef *gqlparser.Definition, inlineFragment *gqlparser.InlineFragment)
	fragmentSpread        []func(walker *Walker, parentDef *gqlparser.Definition, fragmentDef *gqlparser.FragmentDefinition, fragmentSpread *gqlparser.FragmentSpread)
	directive             []func(walker *Walker, parentDef *gqlparser.Definition, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive, location gqlparser.DirectiveLocation)
	directiveList         []func(walker *Walker, parentDef *gqlparser.Definition, directives []gqlparser.Directive, location gqlparser.DirectiveLocation)
	argument              []func(walker *Walker, arg *gqlparser.Argument)
	value                 []func(walker *Walker, valueType gqlparser.Type, def *gqlparser.Definition, value gqlparser.Value)
	variable              []func(walker *Walker, valueType gqlparser.Type, def *gqlparser.Definition, variable gqlparser.VariableDefinition)
}

func (o *Events) OnOperation(f func(walker *Walker, operation *gqlparser.OperationDefinition)) {
	o.operationVisitor = append(o.operationVisitor, f)
}
func (o *Events) OnOperationLeave(f func(walker *Walker, operation *gqlparser.OperationDefinition)) {
	o.operationLeaveVisitor = append(o.operationLeaveVisitor, f)
}
func (o *Events) OnField(f func(walker *Walker, parentDef *gqlparser.Definition, fieldDef *gqlparser.FieldDefinition, field *gqlparser.Field)) {
	o.field = append(o.field, f)
}
func (o *Events) OnFragment(f func(walker *Walker, parentDef *gqlparser.Definition, fragment *gqlparser.FragmentDefinition)) {
	o.fragment = append(o.fragment, f)
}
func (o *Events) OnInlineFragment(f func(walker *Walker, parentDef *gqlparser.Definition, inlineFragment *gqlparser.InlineFragment)) {
	o.inlineFragment = append(o.inlineFragment, f)
}
func (o *Events) OnFragmentSpread(f func(walker *Walker, parentDef *gqlparser.Definition, fragmentDef *gqlparser.FragmentDefinition, fragmentSpread *gqlparser.FragmentSpread)) {
	o.fragmentSpread = append(o.fragmentSpread, f)
}
func (o *Events) OnDirective(f func(walker *Walker, parentDef *gqlparser.Definition, directiveDef *gqlparser.DirectiveDefinition, directive *gqlparser.Directive, location gqlparser.DirectiveLocation)) {
	o.directive = append(o.directive, f)
}
func (o *Events) OnDirectiveList(f func(walker *Walker, parentDef *gqlparser.Definition, directives []gqlparser.Directive, location gqlparser.DirectiveLocation)) {
	o.directiveList = append(o.directiveList, f)
}
func (o *Events) OnArgument(f func(walker *Walker, arg *gqlparser.Argument)) {
	o.argument = append(o.argument, f)
}
func (o *Events) OnValue(f func(walker *Walker, valueType gqlparser.Type, def *gqlparser.Definition, value gqlparser.Value)) {
	o.value = append(o.value, f)
}
func (o *Events) OnVariable(f func(walker *Walker, valueType gqlparser.Type, def *gqlparser.Definition, variable gqlparser.VariableDefinition)) {
	o.variable = append(o.variable, f)
}

func Walk(schema *gqlparser.Schema, document *gqlparser.QueryDocument, observers *Events) {
	w := Walker{
		Observers: observers,
		Schema:    schema,
		Document:  document,
	}
	w.walk()
}

type Walker struct {
	Context   context.Context
	Observers *Events
	Schema    *gqlparser.Schema
	Document  *gqlparser.QueryDocument

	validatedFragmentSpreads map[string]bool
}

func (w *Walker) walk() {
	for _, child := range w.Document.Operations {
		w.validatedFragmentSpreads = make(map[string]bool)
		w.walkOperation(&child)
	}
	for _, child := range w.Document.Fragments {
		w.validatedFragmentSpreads = make(map[string]bool)
		w.walkFragment(&child)
	}
}

func (w *Walker) walkOperation(operation *gqlparser.OperationDefinition) {
	for _, v := range w.Observers.operationVisitor {
		v(w, operation)
	}

	var def *gqlparser.Definition
	var loc gqlparser.DirectiveLocation
	switch operation.Operation {
	case gqlparser.Query, "":
		def = w.Schema.Query
		loc = gqlparser.LocationQuery
	case gqlparser.Mutation:
		def = w.Schema.Mutation
		loc = gqlparser.LocationMutation
	case gqlparser.Subscription:
		def = w.Schema.Subscription
		loc = gqlparser.LocationSubscription
	}

	w.walkDirectives(def, operation.Directives, loc)

	for _, varDef := range operation.VariableDefinitions {
		typeDef := w.Schema.Types[varDef.Type.Name()]
		for _, v := range w.Observers.variable {
			v(w, varDef.Type, typeDef, varDef)
		}
		if varDef.DefaultValue != nil {
			w.walkValue(varDef.Type, varDef.DefaultValue)
		}
	}

	for _, v := range operation.SelectionSet {
		w.walkSelection(def, v)
	}

	for _, v := range w.Observers.operationLeaveVisitor {
		v(w, operation)
	}
}

func (w *Walker) walkFragment(it *gqlparser.FragmentDefinition) {
	parentDef := w.Schema.Types[it.TypeCondition.Name()]

	w.walkDirectives(parentDef, it.Directives, gqlparser.LocationFragmentDefinition)

	for _, v := range w.Observers.fragment {
		v(w, parentDef, it)
	}

	for _, child := range it.SelectionSet {
		w.walkSelection(parentDef, child)
	}
}

func (w *Walker) walkDirectives(parentDef *gqlparser.Definition, directives []gqlparser.Directive, location gqlparser.DirectiveLocation) {
	for _, v := range w.Observers.directiveList {
		v(w, parentDef, directives, location)
	}

	for _, dir := range directives {
		def := w.Schema.Directives[dir.Name]
		for _, v := range w.Observers.directive {
			v(w, parentDef, def, &dir, location)
		}

		for _, arg := range dir.Arguments {
			var argDef *gqlparser.FieldDefinition
			if def != nil {
				argDef = def.Arguments.ForName(arg.Name)
			}

			w.walkArgument(argDef, &arg)
		}
	}
}

func (w *Walker) walkValue(valueType gqlparser.Type, value gqlparser.Value) {
	var def *gqlparser.Definition
	if valueType != nil {
		def = w.Schema.Types[valueType.Name()]
	}

	for _, v := range w.Observers.value {
		v(w, valueType, def, value)
	}

	if obj, isObj := value.(gqlparser.ObjectValue); isObj {
		for _, v := range obj {
			var fieldType gqlparser.Type
			if def != nil {
				fieldDef := def.Field(v.Name)
				if fieldDef != nil {
					fieldType = fieldDef.Type
				}
			}
			w.walkValue(fieldType, v.Value)
		}
	}
}

func (w *Walker) walkArgument(argDef *gqlparser.FieldDefinition, arg *gqlparser.Argument) {
	for _, v := range w.Observers.argument {
		v(w, arg)
	}

	var argType gqlparser.Type
	if argDef != nil {
		argType = argDef.Type
	}

	w.walkValue(argType, arg.Value)

}

func (w *Walker) walkSelection(parentDef *gqlparser.Definition, it gqlparser.Selection) {
	switch it := it.(type) {
	case gqlparser.Field:
		var def *gqlparser.FieldDefinition
		if it.Name == "__typename" {
			def = &gqlparser.FieldDefinition{
				Name: "__typename",
				Type: gqlparser.NamedType("String"),
			}
		} else if parentDef != nil {
			def = parentDef.Field(it.Name)
		}

		for _, v := range w.Observers.field {
			v(w, parentDef, def, &it)
		}

		var nextParentDef *gqlparser.Definition
		if def != nil {
			nextParentDef = w.Schema.Types[def.Type.Name()]
		}

		for _, arg := range it.Arguments {
			var argDef *gqlparser.FieldDefinition
			if def != nil {
				argDef = def.Arguments.ForName(arg.Name)
			}

			w.walkArgument(argDef, &arg)
		}

		for _, sel := range it.SelectionSet {
			w.walkSelection(nextParentDef, sel)
		}

		w.walkDirectives(nextParentDef, it.Directives, gqlparser.LocationField)

	case gqlparser.InlineFragment:
		for _, v := range w.Observers.inlineFragment {
			v(w, parentDef, &it)
		}

		var nextParentDef *gqlparser.Definition
		if it.TypeCondition.Name() != "" {
			nextParentDef = w.Schema.Types[it.TypeCondition.Name()]
		}

		w.walkDirectives(nextParentDef, it.Directives, gqlparser.LocationInlineFragment)

		for _, sel := range it.SelectionSet {
			w.walkSelection(nextParentDef, sel)
		}

	case gqlparser.FragmentSpread:
		def := w.Document.GetFragment(it.Name)

		for _, v := range w.Observers.fragmentSpread {
			v(w, parentDef, def, &it)
		}

		var nextParentDef *gqlparser.Definition
		if def != nil {
			nextParentDef = w.Schema.Types[def.TypeCondition.Name()]
		}

		w.walkDirectives(nextParentDef, it.Directives, gqlparser.LocationFragmentSpread)

		if def != nil && !w.validatedFragmentSpreads[def.Name] {
			// prevent inifinite recursion
			w.validatedFragmentSpreads[def.Name] = true

			for _, sel := range def.SelectionSet {
				w.walkSelection(nextParentDef, sel)
			}
		}

	default:
		panic(fmt.Errorf("unsupported %T", it))

	}
}

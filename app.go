package protomesh

import (
	"fmt"
	"reflect"

	"github.com/jedib0t/go-pretty/v6/list"
)

type App interface {
	// Config() ConfigSource
	Log() Logger
	Close()
}

type DependencyInjector[D any] interface {
	InjectApp(App)
	Dependency() D
}

type AppDependency interface {
	Attach(app App, dep interface{})
}

type Injector[D any] struct {
	app App
	dep D
}

func NewInjector[D any](app App, dep D) *Injector[D] {
	return &Injector[D]{app, dep}
}

func (a *Injector[D]) Attach(app App, dep any) {
	a.app = app
	a.dep = dep.(D)
}

func (a *Injector[D]) Dependency() D {
	return a.dep
}

// func (a *Injector[Dependency]) Config() ConfigSource {
// 	return a.app.Config()
// }

func (a *Injector[Dependency]) Log() Logger {
	return a.app.Log()
}

func (a *Injector[Dependency]) Close() {
	a.app.Close()
}

func Inject[D any](app App, dep D) {
	inject(app, dep, false, nil)
}

func InjectAndPrint[D any](app App, dep D) {
	inject(app, dep, true, nil)
}

func inject[D any](app App, dep D, print bool, lw list.Writer) {

	depVal := reflect.ValueOf(dep)
	appDep := reflect.TypeOf((*AppDependency)(nil)).Elem()

	if print {
		lw = list.NewWriter()
		lw.SetStyle(list.StyleBulletSquare)
	}

	if depVal.Kind() == reflect.Ptr && depVal.Elem().Kind() == reflect.Struct {

		depEl := depVal.Elem()
		depType := reflect.TypeOf(dep)

		for i := 0; i < depEl.NumField(); i++ {

			fieldVal := depEl.Field(i)

			if fieldVal.Type().Implements(appDep) && fieldVal.Kind() == reflect.Ptr {

				if fieldVal.IsNil() {
					fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
				}

				fieldEl := fieldVal.Elem()

				appInj := fieldEl.FieldByName("Injector")
				if !fieldEl.IsValid() || appInj.Kind() != reflect.Ptr {
					continue
				}

				if appInj.IsZero() {
					appInj.Set(reflect.New(appInj.Type().Elem()))
				}

				if lw != nil {
					lw.AppendItem(fmt.Sprintf("%s\n[%s ---> %s]\n", depType.Elem().Field(i).Name, depType.String(), fieldVal.Type().String()))
					lw.Indent()
				}

				appInst := appInj.Interface()

				appDep := appInst.(AppDependency)

				appDep.Attach(app, dep)

				fieldInst := fieldVal.Interface()

				inject(fieldInst.(App), fieldInst, false, lw)

				if lw != nil {
					lw.UnIndent()
				}

			}

		}

	}

	if print {
		fmt.Println("Dependency hierarchy:")
		fmt.Println(lw.Render())
	}

}

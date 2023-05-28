package graviflow

import (
	"fmt"
	"reflect"

	"github.com/jedib0t/go-pretty/v6/list"
)

type App interface {
	Config() ConfigSource
	Log() Logger
	Close()
}

type DependencyInjector[Dependency any] interface {
	InjectApp(App)
	Dependency() Dependency
}

type AppDependency interface {
	Attach(app App, dep interface{})
}

type AppInjector[Dependency any] struct {
	app App
	dep Dependency
}

func (a *AppInjector[Dependency]) Attach(app App, dep any) {
	a.app = app
	a.dep = dep.(Dependency)
}

func (a *AppInjector[Dependency]) Dependency() Dependency {
	return a.dep
}

func (a *AppInjector[Dependency]) Config() ConfigSource {
	return a.app.Config()
}

func (a *AppInjector[Dependency]) Log() Logger {
	return a.app.Log()
}

func (a *AppInjector[Dependency]) Close() {
	a.app.Close()
}

func InjectApp[D any](app App, dep D) {
	injectApp(app, dep, false, nil)
}

func InjectAppAndPrint[D any](app App, dep D) {
	injectApp(app, dep, true, nil)
}

func injectApp[D any](app App, dep D, print bool, lw list.Writer) {

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

				appInj := fieldEl.FieldByName("AppInjector")
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

				injectApp(fieldInst.(App), fieldInst, false, lw)

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

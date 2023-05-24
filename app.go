package graviflow

import "reflect"

type App[Dependency any] interface {
	Config() ConfigSource
	Log() Logger
	Dependency() Dependency
	Close()
}

type DependencyInjector[Dependency any] interface {
	InjectApp(App[Dependency])
	Dependency() Dependency
}

type AppDependency[Dependency any] interface {
	Attach(app App[Dependency])
}

type AppInjector[Dependency any] struct {
	App[Dependency]
}

func (a AppInjector[Dependency]) Attach(app App[Dependency]) {
	a.App = app
}

func InjectApp[Dependency any](app App[Dependency], dep any) {

	depVal := reflect.ValueOf(dep)
	depType := reflect.TypeOf((*AppDependency[Dependency])(nil)).Elem()

	if depVal.Kind() == reflect.Ptr && depVal.Elem().Kind() == reflect.Struct {

		depEl := depVal.Elem()

		for i := 0; i < depEl.NumField(); i++ {

			if depEl.Type().Implements(depType) {

				if depEl.IsNil() {
					depEl.Set(reflect.New(depEl.Type()))
				}

				appDep := depEl.Interface().(AppDependency[Dependency])

				appDep.Attach(app)

				InjectApp(app, appDep)

			}

		}

	}

}

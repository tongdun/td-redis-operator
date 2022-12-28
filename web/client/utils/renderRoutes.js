import React, { Suspense, useEffect, Fragment } from 'react';
import { Switch, Route, Redirect } from 'react-router-dom';
import { Card } from 'antd';
import LoadingPage from '@/components/LoadingPage';

const RouteInstanceMap = {
  get(key) {
    // eslint-disable-next-line
    return key._routeInternalComponent;
  },
  has(key) {
    // eslint-disable-next-line
    return key._routeInternalComponent !== undefined;
  },
  set(key, value) {
    // eslint-disable-next-line
    key._routeInternalComponent = value;
  },
};

const RouteWithProps = ({ path, exact, strict, render, location, ...rest }) => (
  <Route
    path={path}
    exact={exact}
    strict={strict}
    location={location}
    render={props => render({ ...props, ...rest })}
  />
);

function withRoutes(route, extraProps = {}) {
  if (RouteInstanceMap.has(route)) {
    return RouteInstanceMap.get(route);
  }

  const { Routes } = route;
  let len = Routes.length - 1;
  let Component = args => {
    const { render, ...props } = args;
    return render(props);
  };
  while (len >= 0) {
    const AuthRoute = Routes[len];
    const OldComponent = Component;
    Component = props => (
      <AuthRoute {...props} {...extraProps}>
        <OldComponent {...props} {...extraProps} />
      </AuthRoute>
    );
    len -= 1;
  }

  const ret = args => {
    const { render, ...rest } = args;
    return (
      <Fragment>
        <RouteWithProps {...rest} render={props => <Component {...props} route={route} render={render} />} />
      </Fragment>
    );
  };
  RouteInstanceMap.set(route, ret);
  return ret;
}

const wrapperComponent = Component => props => {
  return (
    <Suspense fallback={<LoadingPage />}>
      <Component {...props} />
    </Suspense>
  );
};

export default function renderRoutes(routes, extraProps = {}, switchProps = {}) {
  if (!Array.isArray(routes)) {
    return null;
  }

  const { roles = [] } = extraProps;

  return (
    <Switch {...switchProps}>
      {routes.map((route, i) => {
        if (route.redirect) {
          return (
            <Redirect
              key={route.key || i}
              from={route.path}
              to={route.redirect}
              exact={route.exact}
              strict={route.strict}
            />
          );
        }

        const RouteRoute = route.Routes ? withRoutes(route, extraProps) : RouteWithProps;

        return (
          <RouteRoute
            key={route.key || i}
            path={route.path}
            exact={route.exact}
            strict={route.strict}
            render={props => {
              const childRoutes = renderRoutes(
                route.routes,
                { ...extraProps },
                {
                  location: props.location,
                },
              );

              if (route.component) {
                const Component = wrapperComponent(route.component);

                const content = (
                  <Component {...props} {...extraProps} route={route}>
                    {childRoutes}
                  </Component>
                );

                if (route.path === '/') {
                  return content;
                }

                return route.layout === false ? (
                  <div style={{ margin: 20 }}>{content}</div>
                ) : (
                  <Card
                    style={{ margin: 20, borderRadius: 3 }}
                    bodyStyle={{ padding: 0, backgroundColor: '#f0f2f5' }}
                    bordered={false}>
                    {content}
                  </Card>
                );
              }

              return childRoutes;
            }}
          />
        );
      })}
      <Redirect from="*" to="/404" />
    </Switch>
  );
}

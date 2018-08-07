import 'whatwg-fetch';
let _ = require('lodash');

let DEFINE = require('./Define');

let HttpUtil = {
    getHost: function () {
        return DEFINE.HOST;
    },
    createRoute: function (route, params) {
        let a = new RegExp(DEFINE.Regex.url);
        var url = '';
        if (a.exec(route)) {
            url = route;
        } else {
            url = 'https://' + HttpUtil.getHost() + '/';
            url = url + route;
        }
        var getParams = {};
        //create params
        for (var param in params) {
            if (url.search(':' + param) !== -1) {
                url = url.replace(':' + param, params[param]);
            } else {
                getParams[param] = params[param];
            }
        }
        return [url, getParams];
    },
    HTTPGet: function (route, params, callback, catchCallback, extra) {
        var [url, getParams] = HttpUtil.createRoute(route, params);
        //create query
        var queryParams = '?';
        for (var param in getParams) {
            queryParams = queryParams + param + '=' + getParams[param] + '&';
        }
        queryParams = queryParams.substr(0, queryParams.length - 1);
        url = url + queryParams;
        HttpUtil.HttpRequest('GET', url, null, callback, catchCallback, extra);
    },
    HTTPPost: function (route, params, callback, catchCallback, extra) {
        var [url, getParams] = HttpUtil.createRoute(route, params);
        //create query
        var bodyparams = {};
        for (var param in getParams) {
            bodyparams[param] = getParams[param];
        }
        HttpUtil.HttpRequest('POST', url, bodyparams, callback, catchCallback, extra);
    },
    HTTPPut: function (route, params, callback, catchCallback, extra) {
        var [url, getParams] = HttpUtil.createRoute(route, params);
        //create query
        var bodyparams = {};
        for (var param in getParams) {
            bodyparams[param] = getParams[param];
        }
        HttpUtil.HttpRequest('Put', url, bodyparams, callback, catchCallback, extra);
    },
    HTTPDel: function (route, params, callback, catchCallback, extra) {
        var [url, getParams] = HttpUtil.createRoute(route, params);
        //create query
        var bodyparams = {};
        for (var param in getParams) {
            bodyparams[param] = getParams[param];
        }
        HttpUtil.HttpRequest('Delete', url, bodyparams, callback, catchCallback, extra);
    },
    HttpRequest: function (type, url, params, callback, catchCallback, extra) {
        let fetchParam = {};
        if (type == 'GET') {
            fetchParam = {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
            };
        } else {
            fetchParam = {
                method: type,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(params)
            };
        }
        var status = 200;
        let shouldRet = false;
        fetch(url, fetchParam)
            .then((response) => {
                status = response.status;
                switch (response.status) {
                    case 200: {
                        // var type = response.headers.get('Content-Type');
                        // if (type == null) {
                        //     shouldRet = true;
                        //     callback(null, extra);
                        //     break;
                        // }
                        // if (type.search('application/json') !== -1) {
                        return response.json();
                        // // } else if (type.search('text') !== -1) {
                        // //     return response.text();
                        // // } else {
                        //     shouldRet = true;
                        //     callback(null, extra);
                        // }
                        break;
                    }
                    case 204: {
                        shouldRet = true;
                        callback(null, extra);
                        break;
                    }
                    default: {
                        return response.json();
                    }
                }
            }).then((responseJson) => {
                if (shouldRet) {
                    return;
                }
                if (status == 200 || status == 204) {
                    callback(responseJson, extra);
                }
                else {
                    console.log(responseJson);
                    if (_.isFunction(catchCallback)) {
                        catchCallback(responseJson, extra);
                    }
                }
            })
            .catch((error) => {
                alert(error);
            });
    }
};

module.exports = HttpUtil;
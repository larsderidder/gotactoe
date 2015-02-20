var app = angular.module('gotactoe', ['ngRoute', 'boardControllers']);

app.config(['$routeProvider', function($routeProvider) {
    $routeProvider
    .when('/game', {
        templateUrl: 'game.html',
        controller: 'BoardCtl'
    })
    .when('/about', {
        templateUrl: 'about.html',
    })
    .otherwise({
        redirectTo: '/game'
    })
}]);

app.filter('reverse', function() {
    return function(items) {
        return items.slice().reverse();
    };
});

angular.module('gotactoe.services', [])

.service('connection', ['$location', 'log', '$rootScope', function($location, log, $rootScope) {
    this.open = function() {
        this.conn = new ReconnectingWebSocket("ws://" + location.host + "/ws");

        this.conn.onclose = function(e) {
            log.logMessage("DISCONNECTED - We'll retry in a sec");
        };

        this.conn.onopen = function(e) {
            log.logMessage("CONNECTED");
        };

        // called when a message is received from the server
        this.conn.onmessage = function(e) {
            var data = angular.fromJson(e.data)
            $rootScope.$broadcast('message', data)
        };
    }

    this.send = function(player, x, y) {
        json = { "Player": player, "X": x, "Y": y }
        this.conn.send(JSON.stringify(json));
    };

    this.close = function(){
        this.conn.close();
    };
}])
.service('log', ['$rootScope', function($rootScope) {
    this.getMessageObject = function(msg) {
        return { "datetime": new Date(), "msg": msg };
    }

    this.logMessage = function(msg) {
        $rootScope.$broadcast('logmsg', this.getMessageObject(msg));
    }

    this.logPlayerMessage = function(msg, player) {
        obj = this.getMessageObject(msg);
        obj.player = player;
        $rootScope.$broadcast('logmsg', obj);
    }
}]);

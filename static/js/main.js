var app = angular.module('gotactoe', ['ngTouch']);

app.filter('reverse', function() {
    return function(items) {
        return items.slice().reverse();
    };
});

app.controller("BoardCtl", function($scope) {
    $scope.messages = [];
    $scope.board = {};
    $scope.player = '';
    $scope.other = '';
    $scope.outcome = '';
    $scope.xPlayers = -1;
    $scope.oPlayers = -1;
    $scope.turn = '';
    messageHandlers = {};

    var conn = new ReconnectingWebSocket("ws://" + location.host + "/ws");

    conn.onclose = function(e) {
        $scope.$apply(function() {
            logMessage("DISCONNECTED - We'll retry in a sec");
        });
    };

    conn.onopen = function(e) {
        $scope.$apply(function() {
            logMessage("CONNECTED");
        })
    };

    // called when a message is received from the server
    conn.onmessage = function(e) {
        $scope.$apply(function() {
            var data = angular.fromJson(e.data)
            if (data.Type in messageHandlers) {
                messageHandlers[data.Type](data)
            } else {
                logMessage(data);
            }
        });
    };

    $scope.vote = function vote(x, y) {
        if ($scope.voted) {
            logMessage("You already voted!");
        } else if ($scope.player != $scope.turn) {
            logMessage("Not your turn!");
        } else if ($scope.board[y][x].Player != "") {
            logMessage("Ocupado!");
        } else {
            $scope.board[y][x].voted = true;
            $scope.voted = true;
            send(x, y);
        }
    };

    function getMessageObject(msg) {
        return { "datetime": new Date(), "msg": msg };
    }

    function logMessage(msg) {
        $scope.messages.push(getMessageObject(msg));
    }

    function logPlayerMessage(msg, player) {
        obj = getMessageObject(msg);
        obj.player = player;
        $scope.messages.push(obj);
    }

    function send(x, y) {
        json = { "Player": $scope.player, "X": x, "Y": y }
        conn.send(JSON.stringify(json));
    };

    function updateBoard(data) {
        $scope.board = data.Fields;
        $scope.turn = data.Turn;
    };

    messageHandlers['board'] = function(data) {
        updateBoard(data);
        $scope.voted = false;
        $scope.outcome = '';
    };

    messageHandlers['register'] = function(data) {
        $scope.player = data.Player;
        logPlayerMessage("You are " + $scope.player + "!", $scope.player);
        if (data.Player == 'X') {
            $scope.other = 'O';
        } else {
            $scope.other = 'X';
        }
    };

    messageHandlers['stats'] = function(data) {
        $scope.xPlayers = data.XPlayers;
        $scope.oPlayers = data.OPlayers;
    };

    messageHandlers['outcome'] = function(data) {
        if (data.Outcome == "tie") {
            msg = "It's a tie!"
        } else {
            msg = "The winner is " + data.Outcome + "!"
        }
        $scope.outcome = data.Outcome;
        logPlayerMessage(msg, data.Outcome);
    }
})
.directive('gttField', function() {
    return {
        restrict: 'A',
        link: function link(scope, element, attrs) {
            attrs.$observe('coord', function(value) {
                var coords = value.split("-");
                var x = coords[0];
                var y = coords[1];
                var rawEl = element.get(0);
                var ctx = rawEl.getContext("2d");
                var player = scope.board[y][x].Player;
                if (player == 'X') {
                    ctx.beginPath();
                    drawX(ctx);
                } else if (player == 'O') {
                    ctx.beginPath();
                    drawO(ctx, rawEl.width, rawEl.height);
                }
                if (player) {
                    ctx.lineWidth = 10;
                    if (scope.player == player) {
                        ctx.strokeStyle = '#659f13';
                    } else {
                        ctx.strokeStyle = '#d85030';
                    }
                    ctx.stroke();
                    ctx.closePath();
                }
            });
        }
    }
})
.directive('board', function() {
    return {
        restrict: 'E',
        templateUrl: 'board.html'
    }
});

var app = angular.module("gotactoe", []);

app.filter('reverse', function() {
    return function(items) {
        return items.slice().reverse();
    };
});

app.controller("BoardCtl", function($scope, $http) {
    $scope.messages = [];
    $scope.board = {};
    $scope.player = '';
    $scope.turn = '';

    var conn = new ReconnectingWebSocket("ws://" + location.host + "/ws");
    // called when the server closes the connection
    conn.onclose = function(e) {
        $scope.$apply(function() {
            logMessage("DISCONNECTED - We'll retry in a sec");
        });
    };

    // called when the connection to the server is made
    conn.onopen = function(e) {
        $scope.$apply(function() {
            logMessage("CONNECTED");
            var players = ['O', 'X'];
            $scope.player = players[Math.floor(Math.random()*players.length)];
            logMessage("You are " + $scope.player + "!");
        })
    };

    // called when a message is received from the server
    conn.onmessage = function(e) {
        $scope.$apply(function() {
            var data = angular.fromJson(e.data)
			console.log(data);
            if (data.Type == "board") {
                updateBoard(data);
                $scope.voted = false;
            } else if (data.Type == "outcome") {
                if (data.Outcome == "tie") {
                    msg = "It's a tie!"
                } else {
                    msg = "The winner is " + data.Outcome + "!"
                }
                logMessage(msg);
            } else {
                logMessage(msg);
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

    function logMessage(msg) {
        $scope.messages.push({ "datetime": new Date(), "msg": msg });
    }

    function send(x, y) {
        json = { "Player": $scope.player, "X": x, "Y": y }
        conn.send(JSON.stringify(json));
    };

    function updateBoard(data) {
        $scope.board = data.Fields;
        $scope.turn = data.Turn;
    };
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
                    ctx.moveTo(20, 20);
                    ctx.lineTo(80, 80);
                    ctx.moveTo(80, 20);
                    ctx.lineTo(20, 80);
                } else if (player == 'O') {
                    ctx.beginPath();
                    ctx.arc(rawEl.width / 2, rawEl.height / 2, 30, 0, Math.PI*2, true);
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

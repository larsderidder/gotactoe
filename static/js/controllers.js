var boardControllers = angular.module('gotactoe.board', ['ngTouch', 'gotactoe.services']);

boardControllers.controller("BoardCtl", ['$scope', 'connection', 'log', function($scope, connection, log) {
    $scope.messages = [];
    $scope.board = {};
    $scope.player = '';
    $scope.other = '';
    $scope.outcome = '';
    $scope.xPlayers = -1;
    $scope.oPlayers = -1;
    $scope.turn = '';
    messageHandlers = {};

    connection.open();
    $scope.$on('$locationChangeStart', function(event) {
        connection.close();
    });

    $scope.vote = function vote(x, y) {
        if ($scope.voted) {
            log.logMessage("You already voted!");
        } else if ($scope.player != $scope.turn) {
            log.logMessage("Not your turn!");
        } else if ($scope.board[y][x].Player != "") {
            log.logMessage("Ocupado!");
        } else {
            $scope.board[y][x].voted = true;
            $scope.voted = true;
            connection.send($scope.player, x, y);
        }
    };

    $scope.$on('logmsg', function(event, msg) {
        $scope.messages.push(msg);
    });

    function updateBoard(data) {
        $scope.board = data.Fields;
        $scope.turn = data.Turn;
    };

    $scope.$on('message', function(event, data) {
        if (data.Type in messageHandlers) {
            messageHandlers[data.Type](data)
        } else {
            log.logMessage(data);
        }
    });

    messageHandlers['board'] = function(data) {
        $scope.$apply(function() {
            updateBoard(data);
            $scope.voted = false;
            $scope.outcome = '';
        });
    };

    messageHandlers['register'] = function(data) {
        $scope.player = data.Player;
        log.logPlayerMessage("You are " + $scope.player + "!", $scope.player);
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
        log.logPlayerMessage(msg, data.Outcome);
    }
}])
.directive('gttField', function() {
    return {
        restrict: 'A',
        link: function link(scope, element, attrs) {
            attrs.$observe('player', function(value) {
                var rawEl = element.get(0);
                var ctx = rawEl.getContext("2d");
                var player = value;
                if (player) {
                    ctx.beginPath();
                    if (player == 'X') {
                        drawX(ctx);
                    } else if (player == 'O') {
                        drawO(ctx, rawEl.width, rawEl.height);
                    }
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

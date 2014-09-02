function drawX(ctx) {
    ctx.moveTo(20, 20);
    ctx.lineTo(80, 80);
    ctx.moveTo(80, 20);
    ctx.lineTo(20, 80);
};

function drawO(ctx, width, height) {
    ctx.arc(width / 2, height / 2, 30, 0, Math.PI*2, true);
};

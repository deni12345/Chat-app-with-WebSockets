let websocket;
window.onload = function() {
    websocket = new WebSocket("ws://" + window.location.host + "/ws");
    console.log(websocket);
    websocket.addEventListener("message", function(e) {
        var msg = JSON.parse(e.data);
        let chat_messages = document.getElementById("chat-messages");
        let user =
            `<div class="chip">
                    <img src="${gravatarURL(msg.email)}"> ${
        msg.username
      } </div>` +
            msg.message +
            "<br/>";
        chat_messages.innerHTML += user;
        chat_messages.scrollTop = chat_messages.scrollHeight;
    });
};

function Send() {
    let newMess = document.getElementById("newMsg");
    let email = document.getElementById("email").value;
    let username = document.getElementById("username").value;
    var isAccepted = validfield();
    if (isAccepted) {
        websocket.send(
            JSON.stringify({
                email: email,
                username: username,
                message: newMess.value, // Strip out html
            })
        );
    } else {
        window.loc;
    }

    newMess.value = "";
}

function validfield() {
    let newMess = document.getElementById("newMsg").value;
    let email = document.getElementById("email").value;
    let username = document.getElementById("username").value;
    if (email == "" || username == "" || newMess == "") return false;
    return true;
}

function gravatarURL(email) {
    return "http://www.gravatar.com/avatar/" + CryptoJS.MD5(email);
}
document.addEventListener("DOMContentLoaded", () => {
    const container = document.querySelector(".container");
    const roomId = container.dataset.roomId;
    const body = document.querySelector("body");

    let socket;

    if (body && body.dataset.ws) {
        const wsUrl = body.dataset.ws;
        socket = new WebSocket(wsUrl);
    } else {
        console.error("Atrybut data-ws nie został znaleziony w elemencie <body>.");
    }

    socket.onopen = () => {
        socket.send(
            JSON.stringify({
                type: "check",
                room_id: roomId,
            })
        );
    };

    const passwordInput = document.getElementById("roomPassword");

    document.getElementById("joinRoomForm")?.addEventListener("submit", (e) => {
        e.preventDefault();

        const userName = document.getElementById("userName").value;
        const roomPassword = document.getElementById("roomPassword").value;
        sessionStorage.setItem("roomPassword", roomPassword);
        socket.send(
            JSON.stringify({
                type: "join",
                room_id: roomId,
                user_name: userName,
                password: roomPassword,
            })
        );
    });

    socket.onmessage = (event) => {
        const data = JSON.parse(event.data);
        switch (data.type) {
            case "joinedRoom":
                const roomPassword = document.getElementById("roomPassword").value;
                sessionStorage.setItem("sessionUUID", data.user.id);
                sessionStorage.setItem("sessionUser", data.user.name);
                sessionStorage.setItem("roomPassword", roomPassword);
                window.location.href = `/voting/${data.room.id}`;
                break;
            case "info":
                if (!data.has_password) {
                    passwordInput.style.display = "none";
                }
                break;
        }
    };

    socket.onclose = (event) => {
        console.log("WebSocket closed:", event.code, event.reason);
    };
});

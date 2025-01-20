document.getElementById("createRoomForm")?.addEventListener("submit", (e) => {
    e.preventDefault();

    const roomName = document.getElementById("roomName").value;
    const creatorName = document.getElementById("creatorName").value;
    const roomPassword = document.getElementById("roomPassword").value;
    const roomMethod = document.getElementById("roomMethod").value;

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
                type: "create",
                room_name : roomName,
                user_name: creatorName,
                password: roomPassword,
                room_method: roomMethod,
            })
        );
    };

    socket.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.type === "roomCreated") {
            sessionStorage.setItem("sessionUUID", data.user.id);
            sessionStorage.setItem("sessionUser", data.user.name);
            sessionStorage.setItem("roomPassword", roomPassword);
            window.location.href = `/voting/${data.room.id}`;
        }
    };

    socket.onclose = (event) => {
        console.log("WebSocket closed:", event.code, event.reason);
    };
});

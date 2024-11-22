document.getElementById("createRoomForm")?.addEventListener("submit", (e) => {
    e.preventDefault();

    const roomName = document.getElementById("roomName").value;
    const creatorName = document.getElementById("creatorName").value;
    const roomPassword = document.getElementById("roomPassword").value;

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
                roomName,
                name: creatorName,
                password: roomPassword,
            })
        );
    };

    socket.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.type === "roomCreated") {
            sessionStorage.setItem("userName", data.creatorName);
            sessionStorage.setItem("userId", data.creatorId);
            window.location.href = `/voting-admin/${data.roomId}`;
        }
    };

    socket.onclose = (event) => {
        console.log("WebSocket closed:", event.code, event.reason);
    };
});

document.addEventListener("DOMContentLoaded", () => {
    const container = document.querySelector(".container");
    const roomId = container.dataset.roomId; // Odczyt z atrybutu data-room-id
    const body = document.querySelector("body");

    let socket;

    if (body && body.dataset.ws) {
        const wsUrl = body.dataset.ws;
        socket = new WebSocket(wsUrl);
    } else {
        console.error("Atrybut data-ws nie został znaleziony w elemencie <body>.");
    }

    socket.onopen = () => {
        console.log("WebSocket connected to room:", roomId);
    };

    document.getElementById("joinRoomForm")?.addEventListener("submit", (e) => {
        e.preventDefault();

        const userName = document.getElementById("userName").value;
        const roomPassword = document.getElementById("roomPassword").value;

        socket.send(
            JSON.stringify({
                type: "join",
                roomId: roomId,
                name: userName,
                password: roomPassword,
            })
        );
    });

    socket.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.type === "joinedRoom") {
            console.log("Joined room:", data.roomName);
            // Zapisanie nazwy użytkownika w sessionStorage
            sessionStorage.setItem("userName", data.userName);
            sessionStorage.setItem("userId", data.userId);
            window.location.href = `/voting/${data.roomId}`;
        }
    };

    socket.onclose = (event) => {
        console.log("WebSocket closed:", event.code, event.reason);
    };
});

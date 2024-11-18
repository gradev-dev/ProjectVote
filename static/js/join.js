document.addEventListener("DOMContentLoaded", () => {
    const container = document.querySelector(".container");
    const roomId = container.dataset.roomId; // Odczyt z atrybutu data-room-id
    const socket = new WebSocket("ws://localhost:8080/ws");

    socket.onopen = () => {
        console.log("WebSocket connected to room:", roomId);
    };

    document.getElementById("joinRoomForm")?.addEventListener("submit", (e) => {
        e.preventDefault();

        const userName = document.getElementById("userName").value;
        const roomPassword = document.getElementById("roomPassword").value;

        // Zapisanie nazwy użytkownika w sessionStorage
        sessionStorage.setItem("userName", userName);

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

        if (data.error) {
            alert(data.error);
        } else {
            console.log("Joined room:", data);
            window.location.href = `/voting/${roomId}`;
        }
    };

    socket.onclose = (event) => {
        console.log("WebSocket closed:", event.code, event.reason);
    };
});

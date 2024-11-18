document.getElementById("createRoomForm")?.addEventListener("submit", (e) => {
    e.preventDefault();

    const roomName = document.getElementById("roomName").value;
    const creatorName = document.getElementById("creatorName").value;
    const roomPassword = document.getElementById("roomPassword").value;

    const socket = new WebSocket("ws://localhost:8080/ws");

    // Zapisanie nazwy użytkownika w sessionStorage
    sessionStorage.setItem("userName", creatorName);

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
            console.log("Room created with ID:", data.roomId);
            window.location.href = `/voting/${data.roomId}`;
        }
    };

    socket.onclose = (event) => {
        console.log("WebSocket closed:", event.code, event.reason);
    };
});

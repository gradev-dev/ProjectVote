document.addEventListener("DOMContentLoaded", () => {
    const container = document.querySelector(".container-voting-right");
    const roomId = container.dataset.roomId;

    // Pobierz nazwę użytkownika z sessionStorage
    const userUUID = sessionStorage.getItem("userId");
    const userName = sessionStorage.getItem("userName");

    if (!userUUID) {
        window.location.href = `/join/${roomId}`;
        return;
    }

    const body = document.querySelector("body");

    let socket;

    if (body && body.dataset.ws) {
        const wsUrl = body.dataset.ws;
        socket = new WebSocket(wsUrl);
    } else {
        console.error("Atrybut data-ws nie został znaleziony w elemencie <body>.");
    }

    socket.onopen = () => {
        // Wyślij żądanie dołączenia
        socket.send(JSON.stringify({ type: "join", roomId, name: userName, userId: userUUID }));
    };

    document.getElementById("voteButtons")?.addEventListener("click", (e) => {
        if (e.target.tagName === "BUTTON") {
            const vote = parseInt(e.target.textContent, 10);
            const buttons = document.querySelectorAll('.vote-btn');
            buttons.forEach(btn => btn.classList.remove('vote-selected'));
            e.target.classList.add('vote-selected');
            socket.send(
                JSON.stringify({
                    type: "vote",
                    roomId: roomId,
                    userId: userUUID,
                    vote: vote,
                })
            );
        }
    });

    socket.onmessage = (event) => {
        const data = JSON.parse(event.data);

        switch (data.type) {
            case "joinedRoom":
                document.getElementById("roomName").textContent = data.roomName;
                document.getElementById("userName").textContent = userName;
                break;

            case "update":
                updateParticipants(data.participants, data.reveal, data.reset);
                break;

            case "error":
                alert(data.message);
                break;

            default:
                console.warn(`Unknown message type: ${data.type}`);
        }
    };


    socket.onclose = (event) => {
        console.log("WebSocket closed:", event.code, event.reason);
    };

    const updateParticipants = (participants, reveal, reset) => {
        const list = document.getElementById("participants");
        list.innerHTML = ""; // Wyczyszczenie listy uczestników
        if (reset) {
            const buttons = document.querySelectorAll('.vote-btn');
            buttons.forEach(btn => {
                btn.classList.remove('vote-selected');
                btn.removeAttribute('disabled');
            });
        }

        for (const [uuid, participant] of Object.entries(participants)) {
            // Tworzenie kontenera dla karty
            const card = document.createElement("div");
            card.className = "card";

            // Dodanie obiektu wewnętrznego dla obrotu
            const cardInner = document.createElement("div");
            cardInner.className = "card-inner";

            // Tworzenie przedniej strony karty
            const cardFront = document.createElement("div");
            cardFront.className = "card-face card-front";
            cardFront.textContent = participant.name; // Dodanie imienia uczestnika

            // Tworzenie tylnej strony karty
            const cardBack = document.createElement("div");
            cardBack.className = "card-face card-back";

            // Dodanie głosu do środka karty
            const voteDiv = document.createElement("div");
            voteDiv.className = "vote";
            voteDiv.textContent = reveal ? participant.vote : "?"; // Wyświetlenie głosu lub "?"

            // Dodanie imienia uczestnika na dole
            const nameDiv = document.createElement("div");
            nameDiv.className = "participant-name";
            nameDiv.textContent = participant.name;

            cardBack.appendChild(nameDiv);
            cardBack.appendChild(voteDiv);

            cardInner.appendChild(cardFront);
            cardInner.appendChild(cardBack);
            card.appendChild(cardInner);

            if (reveal) {
                setTimeout(() => card.classList.add("revealed"), 100); // Płynne dodanie klasy z opóźnieniem
                const buttons = document.querySelectorAll('.vote-btn');
                buttons.forEach(btn => btn.setAttribute('disabled', ''));
            }

            list.appendChild(card);
        }
    };
});

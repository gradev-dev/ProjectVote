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
                // Wyświetl nazwę pokoju z backendu
                document.getElementById("roomName").textContent = data.roomName;
                document.getElementById("userName").textContent = userName;
                console.log(`Joined room: ${data.roomName}`);
                // Sprawdź, czy użytkownik jest właścicielem pokoju
                if (data.isOwner) {
                    console.log("You are the owner of this room.");
                    addOwnerControls(socket, roomId, userUUID);
                }
                break;

            case "update":
                // Zaktualizuj listę uczestników
                updateParticipants(data.participants, data.reveal, data.reset);
                break;

            case "error":
                // Obsługa błędu
                console.error("Error received:", data.message);
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
            console.log(uuid);
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

            // Dodanie głosu i imienia do tylnej strony karty
            cardBack.appendChild(nameDiv);
            cardBack.appendChild(voteDiv);

            // Dodanie stron do wnętrza karty
            cardInner.appendChild(cardFront);
            cardInner.appendChild(cardBack);
            card.appendChild(cardInner);

            // Jeśli głosy są odkryte, dodaj klasę "revealed"
            if (reveal) {
                setTimeout(() => card.classList.add("revealed"), 100); // Płynne dodanie klasy z opóźnieniem
                const buttons = document.querySelectorAll('.vote-btn');
                buttons.forEach(btn => btn.setAttribute('disabled', ''));
            }

            // Dodanie karty do listy uczestników
            list.appendChild(card);
        }
    };

    const addOwnerControls = (socket, roomId, userId) => {
        const controlsContainer = document.createElement("div");
        controlsContainer.id = "ownerControls";
        controlsContainer.className = "position-bottom-10";

        // Przycisk "Reveal Votes"
        const revealButton = document.createElement("button");
        revealButton.id = "revealBtn";
        revealButton.textContent = "Reveal Votes";
        revealButton.className = "btn";
        revealButton.addEventListener("click", () => {
            socket.send(JSON.stringify({ type: "reveal", roomId, userId: userId }));
        });

        // Przycisk "Reset Room"
        const resetButton = document.createElement("button");
        resetButton.id = "resetBtn";
        resetButton.textContent = "Reset Room";
        resetButton.className = "btn";
        resetButton.addEventListener("click", () => {
            socket.send(JSON.stringify({ type: "reset", roomId, userId: userId }));
        });

        // Dodaj przyciski do kontenera
        controlsContainer.appendChild(revealButton);
        controlsContainer.appendChild(resetButton);

        // Dodaj kontener do DOM
        container.appendChild(controlsContainer);
    };
});

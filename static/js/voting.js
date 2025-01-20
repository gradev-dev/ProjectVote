const CLIENT_MESSAGE_TYPE_TASK = "task";
const CLIENT_MESSAGE_TYPE_JOIN = "join";
const CLIENT_MESSAGE_TYPE_VOTE = "vote";
const CLIENT_MESSAGE_TYPE_REVEAL = "reveal";
const CLIENT_MESSAGE_TYPE_RESET = "reset";

const SERVER_MESSAGE_TYPE_JOINED_ROOM = "joinedRoom";
const SERVER_MESSAGE_TYPE_UPDATE = "update";
const SERVER_MESSAGE_TYPE_REDIRECT = "redirect";

document.addEventListener("DOMContentLoaded", () => {
    const body = document.querySelector("body"),
        modal = body.querySelector("#modal"),

        leftContainer = body.querySelector(".container-voting-left"),
        history = leftContainer.querySelector("#history"),
        votesHistoryContainer = leftContainer.querySelector("#votes-history"),
        votingControls = leftContainer.querySelector("#voting-controls"),

        rightContainer = body.querySelector(".container-voting-right"),
        average = rightContainer.querySelector("#average"),
        participantsList = rightContainer.querySelector("#participants"),

        project = rightContainer.querySelector(".project-info-admin"),
        projectInput = project.querySelector("#projectInput"),
        projectButton = project.querySelector("#taskInfoBtn"),
        projectStatus = project.querySelector("#projectStatus"),
        projectLoadingIndicator = project.querySelector("#projectLoadingIndicator");

    const userNameSpan = document.getElementById("userName");

    const roomId = rightContainer.dataset.roomId,
         roomMethod = rightContainer.dataset.roomMethod;

    const sessionUUID = sessionStorage.getItem("sessionUUID"),
         sessionUser = sessionStorage.getItem("sessionUser");
    if (!sessionUUID) {
        window.location.href = `/join/${roomId}`;
        return;
    }

    let socket;

    if (body && body.dataset.ws) {
        const wsUrl = body.dataset.ws;
        socket = new WebSocket(wsUrl);
    } else {
        console.error("Atrybut data-ws nie został znaleziony w elemencie <body>.");
    }

    socket.onopen = () => {
        socket.send(JSON.stringify({ type: CLIENT_MESSAGE_TYPE_JOIN, room_id: roomId, user_name: sessionUser, session_uuid: sessionUUID, password: sessionStorage.getItem("roomPassword") }));
    };

    document.getElementById("voteButtons")?.addEventListener("click", (e) => {
        if (e.target.tagName === "BUTTON") {
            const vote = e.target.textContent.toString();
            const buttons = document.querySelectorAll('.vote-btn');
            buttons.forEach(btn => btn.classList.remove('vote-selected'));
            e.target.classList.add('vote-selected');
            socket.send(
                JSON.stringify({
                    type: CLIENT_MESSAGE_TYPE_VOTE,
                    room_id: roomId,
                    user_name: sessionUser,
                    session_uuid: sessionUUID,
                    vote: vote,
                })
            );
        }
    });

    document.getElementById("historyButtons")?.addEventListener("click", (e) => {
        if (e.target.tagName === "BUTTON") {
            if (history.classList.contains("room-history-view")) {
                history.classList.replace("room-history-view", "room-history" );
                e.target.classList.remove("active");
            } else {
                history.classList.replace("room-history", "room-history-view" );
                e.target.classList.add("active");

                votesHistoryContainer.innerHTML = '';

                const votesContainerTop = document.createElement('div');
                votesContainerTop.className = 'task-top';
                votesContainerTop.innerHTML = '<div class="task-top-bar"></div><div></div>';

                const indicator = document.createElement('div');
                indicator.id = 'loadingHistoryIndicator';
                indicator.innerHTML = '<div class="lds-ring"><div></div><div></div><div></div><div></div></div>';

                votesHistoryContainer.appendChild(votesContainerTop);
                votesHistoryContainer.appendChild(indicator);

                fetch(`/history/${roomId}`)
                    .then(response => {
                        if (!response.ok) {
                            indicator.style.display = 'none';
                            const empty = document.createElement('div');
                            empty.className = 'empty-history';
                            votesHistoryContainer.appendChild(empty);
                        }
                        return response.json();
                    })
                    .then(data => {
                        indicator.remove();
                        data.forEach((voteData, index) => {
                            let fibonacci = voteData.fibonacci;
                            let tshirt = voteData.tshirt;

                            const taskDiv = document.createElement('div');
                            taskDiv.className = 'task-container';

                            const headerDiv = document.createElement('div');
                            headerDiv.className = 'task-header';

                            const taskTitle = document.createElement('div');
                            taskTitle.className = 'task-title';
                            taskTitle.textContent = voteData.task;

                            const taskAverage = document.createElement('div');
                            taskAverage.className = 'task-fib';
                            taskAverage.textContent = (roomMethod === 'fibonacci') ? fibonacci : tshirt;

                            const toggleButton = document.createElement('button');
                            toggleButton.innerHTML = '<div><i class="bi bi-chevron-down"></i></div>';
                            toggleButton.dataset.index = index;

                            headerDiv.appendChild(taskAverage);
                            headerDiv.appendChild(taskTitle);
                            headerDiv.appendChild(toggleButton);
                            taskDiv.appendChild(headerDiv);

                            const historParticipantsList = document.createElement('div');
                            historParticipantsList.className = 'history-participants-list';
                            for (const [uuid, participant] of Object.entries(voteData.participants)) {
                                let vote;

                                if (participant.vote === "coffee") {
                                    vote = `<div class="coffee-small"></div>`;
                                } else {
                                    vote = `<p>${participant.vote}</p>`;
                                }

                                const participantDiv = document.createElement('div');
                                participantDiv.className = 'history-participant';
                                participantDiv.innerHTML = `
                                    <span>${participant.name}</span>
                                    ${vote}
                                `;
                                historParticipantsList.appendChild(participantDiv);
                            }

                            taskDiv.appendChild(historParticipantsList);
                            votesHistoryContainer.appendChild(taskDiv);

                            toggleButton.addEventListener('click', () => {
                                const isVisible = historParticipantsList.style.display === 'block';
                                historParticipantsList.style.display = isVisible ? 'none' : 'block';
                                const icon = toggleButton.querySelector('div');
                                icon.classList.toggle('active-task-participant', !isVisible);
                            });
                        });
                    });
            }
        }
    });

    let processedMessages = new Set();
    socket.onmessage = (event) => {
        const data = JSON.parse(event.data);
        console.log(data);

        if (processedMessages.has(data.message_id)) {
            return;
        }

        processedMessages.add(data.message_id);
        switch (data.type) {
            case SERVER_MESSAGE_TYPE_JOINED_ROOM:
                userNameSpan.textContent = data.user.name;
                break;

            case SERVER_MESSAGE_TYPE_UPDATE:
                if (data.user.room_owner) {
                    addOwnerControls(data, socket);
                } else {
                    projectInput.disabled = true;
                }

                updateParticipants(data);
                updateCurrentTask(data.room.current_task, data.user.room_owner, data.room.id, sessionUUID, socket, data.room.reveal);
                break;

            case SERVER_MESSAGE_TYPE_REDIRECT:
                window.location.href = `${data.url}`;
                break;
            default:
                console.warn(`Unknown message type: ${data.type}`);
        }
    };

    socket.onclose = (event) => {
        console.log("WebSocket closed:", event.code, event.reason);
    };

    projectButton.addEventListener('click', () => {
        if (modal.classList.contains('modal-deactive')) {
            modal.classList.replace('modal-deactive', 'modal-active');
            const modalContent = document.querySelector(".modal-content")
            modalContent.innerHTML = '';
            const indicator = document.createElement('div');
            indicator.className = 'modal-indicator';
            indicator.innerHTML = '<div class="lds-ring"><div></div><div></div><div></div><div></div></div>';

            modalContent.appendChild(indicator);

            fetch(`/tasks/detail/${sessionStorage.getItem("lastTask")}`)
                .then(response => {
                    if (!response.ok) {
                        indicator.remove();
                    }
                    return response.json();
                })
                .then(data => {
                    indicator.remove();
                    const title = document.createElement('div');
                    title.className = 'modal-title';
                    title.innerHTML = `<div class="modal-title__title">${data.summary}</div><div class="modal-title__close" onclick="
                        if (modal.classList.contains('modal-active')) {
                            modal.classList.replace('modal-active', 'modal-deactive');
                        }
                    ">X</div>`;

                    const description = document.createElement('div');
                    description.className = 'modal-description';
                    description.innerHTML = `${data.description}`;

                    const comments = document.createElement('div');
                    comments.className = 'modal-comments';
                    comments.innerHTML = `<div class="modal-comments-toolbar"><b>Comments</b></div>`;
                    modalContent.appendChild(title);
                    modalContent.appendChild(description);

                    data.comments.forEach((commentData, index) => {
                        const comment = document.createElement('div');
                        comment.className = 'modal-comment';
                        comment.innerHTML = `<div class="modal-comment-author">${commentData.author}</div><div class="modal-comment-body">${commentData.body}</div>`;
                        comments.appendChild(comment);
                    });

                    modalContent.appendChild(comments);
                });

        } else {
            modal.classList.replace('modal-active', 'modal-deactive');
        }
    });

    const changeTaskInfoButton = (isTask) =>{
        if (isTask && projectButton.classList.contains('task-info-btn')) {
            projectButton.classList.replace('task-info-btn', 'task-info-btn-enabled');
        }
        if (!isTask && projectButton.classList.contains('task-info-btn-enabled')) {
            projectButton.classList.replace('task-info-btn-enabled', 'task-info-btn');
        }
    }

    const updateCurrentTask = (currentTask, isRoomOwner, roomId, sessionUUID, socket, reveal) => {
        if (currentTask !== "") {
            sessionStorage.setItem("lastTask", currentTask);
            changeTaskInfoButton(true);
            initializeVotingControls(reveal, roomId, sessionUUID, socket);

            projectInput.value = currentTask;
            projectInput.disabled = true;
            projectStatus.style.display = 'block';
        } else {
            projectInput.value = "";
            projectStatus.style.display = 'none';
            projectInput.disabled = !isRoomOwner;
        }
    }

    const updateParticipants = (data) => {
        participantsList.innerHTML = "";

        for (const [uuid, participant] of Object.entries(data.room.participants)) {
            const card = document.createElement("div");
            card.className = "card";

            const cardInner = document.createElement("div");
            cardInner.className = "card-inner";

            const cardFront = document.createElement("div");
            cardFront.className = "card-face card-front";

            const cardBack = document.createElement("div");
            cardBack.className = "card-face card-back";

            const voteIconDiv = document.createElement("div");
            voteIconDiv.className = "voteIcon";
            if (participant.vote && participant.vote !== "0") {
                const checkIcon = document.createElement("div");
                checkIcon.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" shape-rendering="geometricPrecision" text-rendering="geometricPrecision" image-rendering="optimizeQuality" fill-rule="evenodd" clip-rule="evenodd" viewBox="0 0 512 512"><path d="M256 0c141.39 0 256 114.62 256 256 0 141.39-114.61 256-256 256C114.62 512 0 397.39 0 256 0 114.62 114.62 0 256 0zm88.56 211.59c4.07-3.97 10.63-3.92 14.63.11s3.95 10.52-.12 14.49l-97.13 94.23c-4.06 3.95-10.58 3.91-14.59-.08l-94.5-94.23c-4.02-4.01-4-10.5.05-14.48 4.05-3.99 10.6-3.97 14.62.04l87.27 87.01 89.77-87.09z"/></svg>'
                checkIcon.className = "checkIcon";
                voteIconDiv.appendChild(checkIcon);
            }

            const nameFrontDiv = document.createElement("div");
            nameFrontDiv.className = "participant-name";
            nameFrontDiv.textContent = participant.name;

            const voteDiv = document.createElement("div");
            voteDiv.className = "vote";
            if (data.room.reveal) {
                if (participant.vote === "coffee") {
                    voteDiv.innerHTML = `
            <?xml version="1.0" encoding="utf-8"?><svg version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px" viewBox="0 0 113.5 122.88" style="enable-background:new 0 0 113.5 122.88" xml:space="preserve"><style type="text/css">.st0{fill-rule:evenodd;clip-rule:evenodd;}</style><g><path class="st0" d="M101.71,56.76h-5.54v16.51h5.55v0.02c1.14,0,2.19-0.49,2.96-1.26l0.14-0.13c0.68-0.75,1.1-1.73,1.1-2.8 l-0.01,0v-0.03h0.01v-8.1h-0.01v-0.02l0.01,0c0-1.15-0.48-2.2-1.24-2.96c-0.76-0.76-1.81-1.24-2.95-1.25v0.01H101.71L101.71,56.76 L101.71,56.76z M31.66,0.61c1.75-1.16,4.08-0.63,5.2,1.17c1.12,1.8,0.61,4.21-1.14,5.37c-3.19,2.11-3.21,3.75-2.23,5.22 c0.84,1.27,2.14,2.66,3.43,4.03c1.79,1.91,3.59,3.82,4.82,6.03c2.93,5.23,2.67,10.43-6.55,15.59c-1.82,1.02-4.11,0.33-5.1-1.55 c-0.99-1.88-0.32-4.24,1.5-5.26c3.85-2.15,4.34-3.66,3.63-4.92c-0.74-1.31-2.22-2.89-3.7-4.47c-1.53-1.63-3.07-3.27-4.26-5.07 C23.83,11.54,23.21,6.2,31.66,0.61L31.66,0.61z M74.07,0.61c1.75-1.16,4.08-0.63,5.2,1.17c1.12,1.8,0.61,4.21-1.14,5.37 c-3.19,2.11-3.2,3.75-2.23,5.22c0.84,1.27,2.14,2.66,3.43,4.03c1.79,1.91,3.59,3.82,4.82,6.03c2.93,5.23,2.68,10.43-6.54,15.59 c-1.82,1.02-4.11,0.33-5.1-1.55c-0.99-1.88-0.32-4.24,1.5-5.26c3.85-2.15,4.33-3.65,3.62-4.92c-0.74-1.31-2.22-2.89-3.7-4.47 c-1.53-1.63-3.07-3.27-4.26-5.07C66.24,11.53,65.61,6.2,74.07,0.61L74.07,0.61z M52.87,0.61c1.75-1.16,4.08-0.63,5.2,1.17 c1.12,1.8,0.61,4.21-1.14,5.37c-3.19,2.11-3.21,3.75-2.23,5.22c0.84,1.27,2.14,2.66,3.43,4.03c1.79,1.91,3.59,3.82,4.82,6.03 c2.93,5.23,2.68,10.43-6.54,15.59c-1.82,1.02-4.11,0.33-5.1-1.55c-0.99-1.88-0.32-4.24,1.5-5.26c3.85-2.15,4.33-3.66,3.62-4.92 c-0.74-1.31-2.22-2.89-3.7-4.47c-1.53-1.63-3.08-3.27-4.26-5.07C45.03,11.54,44.42,6.2,52.87,0.61L52.87,0.61z M1.42,112.34h36.12 c-12.25-6.13-20.72-18.8-20.72-33.37V48.92h74.58v0.39c0.32-0.09,0.66-0.13,1.01-0.13l9.35,0v0.02h0.02 c3.22,0.01,6.14,1.32,8.26,3.44c2.13,2.12,3.46,5.07,3.47,8.31l0.01,0v0.02h-0.01v7.96l0,0.14h0.01v0.03h-0.01v0.02 c-0.01,3.08-1.22,5.9-3.18,7.99c-0.08,0.1-0.17,0.19-0.26,0.28c-2.12,2.12-5.07,3.44-8.32,3.45v0.02l-9.34,0 c-0.36,0-0.72-0.06-1.05-0.15c-0.63,13.84-8.9,25.77-20.67,31.65h38.15c0.78,0,1.42,0.76,1.42,1.7v7.15c0,0.94-0.64,1.69-1.42,1.69 l-107.4,0c-0.78,0-1.42-0.76-1.42-1.69v-7.15C0,113.1,0.64,112.34,1.42,112.34L1.42,112.34L1.42,112.34z"/></g></svg>
        `;
                } else {
                    voteDiv.textContent = participant.vote;
                }
            } else {
                voteDiv.textContent = "?";
            }

            const nameBackDiv = document.createElement("div");
            nameBackDiv.className = "participant-name";
            nameBackDiv.textContent = participant.name;

            cardFront.appendChild(nameFrontDiv);
            cardFront.appendChild(voteIconDiv);

            cardBack.appendChild(nameBackDiv);
            cardBack.appendChild(voteDiv);

            cardInner.appendChild(cardFront);
            cardInner.appendChild(cardBack);
            card.appendChild(cardInner);

            if (data.room.reveal) {
                setTimeout(() => card.classList.add("revealed"), 100); // Płynne dodanie klasy z opóźnieniem
                const buttons = document.querySelectorAll('.vote-btn');
                buttons.forEach(btn => btn.setAttribute('disabled', ''));
            }

            participantsList.appendChild(card);
        }

        if (data.room.current_task !== "") {
            sessionStorage.setItem("lastTask", data.room.current_task);
            changeTaskInfoButton(true);
            initializeVotingControls(data.room.reveal, roomId, sessionUUID, socket);

            projectInput.value = data.room.current_task;
            projectInput.disabled = true;
        } else {
            projectStatus.style.display = 'none';

            if (data.user.room_owner) {
                projectInput.value = '';
                projectInput.disabled = false;
            } else {
                projectInput.value = `${data.room.last_task}`;
                projectInput.disabled = true;
            }
        }

        if (data.room.reset) {
            changeTaskInfoButton(false);

            if (modal.classList.contains("modal-active")) {
                modal.classList.replace("modal-active", "modal-deactive" );
            }

            deleteVotingControls();
            const element = document.querySelector('.average-info');
            if (element) {
                element.remove();
            }

            if (data.user.room_owner) {
                addOwnerControls(data, socket);
            }
        }

        if (data.room.reveal) {

            if (!data.user.room_owner) {
                const averageInfo = document.createElement("div");
                averageInfo.className = "average-info position-bottom-0";

                averageInfo.appendChild(updateAverageVoting());
                average.appendChild(averageInfo);
            }

            switch (roomMethod) {
                case "fibonacci":
                    const votingAverageSpan = document.getElementById("votingAverage");
                    if (votingAverageSpan) {
                        votingAverageSpan.textContent = `${data.voting.average}`;
                    }

                    const votingAverageFibSpan = document.getElementById("votingAverageFib");
                    if (votingAverageFibSpan)  {
                        votingAverageFibSpan.textContent = `${data.voting.fibonacci.toString()}`;
                    }
                   break;
                case "tshirts":
                    const votingAverageTshirtSpan = document.getElementById("votingAverageTshirt");
                    if (votingAverageTshirtSpan)  {
                        votingAverageTshirtSpan.textContent = `${data.voting.tshirt}`;
                    }
                    break;
            }
        }
    };

    projectInput.addEventListener("keydown", function (event) {
        if (event.key === "Enter") {
            const task = this.value.trim();
            projectLoadingIndicator.style.display = 'block';
            if (task.length >= 3) {
                fetch(`/tasks/search/${task}`)
                    .then(response => {
                        if (!response.ok) {
                            if (response.status === 404 || response.status === 500) {
                                response.json().then(data => {
                                    showErrorNotification(data.error);
                                });
                            }
                            projectLoadingIndicator.style.display = 'none';
                            return Promise.reject(new Error('Failed to fetch'));
                        }

                        projectLoadingIndicator.style.display = 'none';
                        socket.send(JSON.stringify({
                            type: CLIENT_MESSAGE_TYPE_TASK,
                            room_id: roomId,
                            user_name: sessionUser,
                            session_uuid: sessionUUID,
                            task_name: task
                        }));
                        projectStatus.style.display = 'block';
                        return Promise.resolve();
                    })
                    .catch(error => {
                        projectStatus.style.display = 'none';
                    });
            }
        }
    });

    const addOwnerControls = (data, socket) => {
        if (data.room.current_task === "") {
            projectInput.placeholder = 'e.g., PSC [press Enter...]';
            projectInput.disabled = false;
            return;
        }

        const element = document.querySelector('.average-info');
        if (element) {
            element.remove();
        }

        const controlsContainer = document.createElement("div");
        controlsContainer.id = "ownerControls";

        const revealButton = document.createElement("button");
        revealButton.id = "revealBtn";
        revealButton.textContent = "Reveal Votes";
        revealButton.className = "btn";
        revealButton.addEventListener("click", () => {
            socket.send(JSON.stringify({ type: CLIENT_MESSAGE_TYPE_REVEAL, room_id: roomId, user_name: sessionUser, session_uuid: sessionUUID }));
        });

        const resetButton = document.createElement("button");
        resetButton.id = "resetBtn";
        resetButton.textContent = "Reset Room";
        resetButton.className = "btn";
        resetButton.addEventListener("click", () => {
            socket.send(JSON.stringify({ type: CLIENT_MESSAGE_TYPE_RESET, room_id: roomId, user_name: sessionUser, session_uuid: sessionUUID }));
        });

        const averageInfo = document.createElement("div");
        averageInfo.className = "average-info position-bottom-0";

        if (data.room.reveal) {
            averageInfo.appendChild(updateAverageVoting());
            controlsContainer.appendChild(resetButton);
            if (roomMethod === 'fibonacci') {
                const averageBtn = document.createElement("button");
                averageBtn.id = "averageBtn";
                averageBtn.textContent = "save on Jira";
                averageBtn.className = "btn";
                averageBtn.addEventListener("click", () => {
                    const fib = document.querySelector("#votingAverageFib");
                    fetch(`/tasks/save`, {
                        method: "PUT",
                        headers: {
                            "Content-Type": "application/json",
                        },
                        body: JSON.stringify({task: sessionStorage.getItem("lastTask"), fib: fib.innerHTML })
                    })
                        .then(response => {
                            if (!response.ok) {
                                throw new Error("Task not found");
                            }
                            return response.json();
                        })
                        .then(data => {
                            averageBtn.classList.remove("btn");
                            averageBtn.classList.add("btn-saved");
                        })
                        .catch(error => {
                            console.error("Error:", error);
                        });

                });

                controlsContainer.appendChild(averageBtn);
            }
        } else {
            controlsContainer.appendChild(revealButton);
        }

        averageInfo.appendChild(controlsContainer);
        average.appendChild(averageInfo);
    };

    function updateAverageVoting() {
        const averageVoting = document.createElement("div");
        averageVoting.className = "average-voting";

        if (roomMethod === 'fibonacci') {
            const averageVotingLeft = document.createElement("div");
            averageVotingLeft.className = "voting-info left-border";
            averageVotingLeft.innerHTML = '<p>Average</p><span id="votingAverage">0.0</span>';

            const averageVotingRight = document.createElement("div");
            averageVotingRight.className = "voting-info";
            averageVotingRight.innerHTML = '<p>Average Fibonacci</p><span id="votingAverageFib">0</span>';

            averageVoting.appendChild(averageVotingLeft);
            averageVoting.appendChild(averageVotingRight);
        } else {
            const averageVotingRight = document.createElement("div");
            averageVotingRight.className = "voting-info";
            averageVotingRight.innerHTML = '<p>Average T-Shirt</p><span id="votingAverageTshirt">0</span>';
            averageVoting.appendChild(averageVotingRight);
        }

        return averageVoting;
    }

    function createVoteButtons(reveal = false, options, roomId, userUUID, socket, tooltips = null) {
        const voteButtons = document.createElement("div");
        voteButtons.id = "voteButtons";
        const voteButtonCoffee = document.createElement("button");
        voteButtonCoffee.className = "vote-btn coffee";
        voteButtonCoffee.innerHTML = `<span class="tooltip">An extraordinary drink that changes the dynamics of reality</span>`
        voteButtonCoffee.setAttribute('data-vote', 'coffee');
        if (reveal) {
            voteButtonCoffee.disabled = true;
        }

        voteButtons.appendChild(voteButtonCoffee);
        fadeInElement(voteButtonCoffee, 0);

        options.forEach((option, index) => {
            const button = document.createElement('button');
            button.className = "vote-btn";
            button.innerText = option;
            button.setAttribute('data-vote', option);
            if (tooltips && tooltips[index]) {
                button.innerHTML = `${option} <span class="tooltip">${tooltips[index]}</span>`
            } else {
                button.innerText = option;
            }

            if (reveal) {
                button.disabled = true;
            }
            voteButtons.appendChild(button);
            fadeInElement(button, (index + 1) * 100);
        });

        voteButtons.addEventListener('click', e => handleVoteClick(e, roomId, userUUID, socket));
        return voteButtons;
    }

    function handleVoteClick(e, roomId, userUUID, socket) {
        if (e.target.tagName === 'BUTTON') {
            let voteValue;
            if (e.target.classList.contains('vote-btn')) {
                voteValue = e.target.getAttribute('data-vote');
            }
            const buttons = document.querySelectorAll('.vote-btn');
            buttons.forEach(btn => btn.classList.remove('vote-selected'));
            e.target.classList.add('vote-selected');
            socket.send(
                JSON.stringify({
                    type: CLIENT_MESSAGE_TYPE_VOTE,
                    room_id: roomId,
                    user_name: sessionUser,
                    session_uuid: sessionUUID,
                    vote: voteValue,
                })
            );
        }
    }

    function fadeInElement(element, delay) {
        element.style.opacity = 0;
        setTimeout(() => {
            let opacity = 0;
            const interval = setInterval(() => {
                if (opacity >= 1) {
                    clearInterval(interval);
                } else {
                    opacity += 0.1;
                    element.style.opacity = opacity;
                }
            }, 50);
        }, delay);
    }

    function deleteVotingControls() {
        const voteButtons = document.getElementById("voteButtons");
        if (voteButtons) {
            voteButtons.remove();
        }
    }

    function initializeVotingControls(reveal = false, roomId, userUUID, socket) {
        const voteButtons = document.getElementById("voteButtons");
        if (voteButtons) {
            return;
        }

        if (roomMethod === "fibonacci") {
            const fibNumbers = [1, 2, 3, 5, 8, 13];
            votingControls.appendChild(createVoteButtons(reveal,fibNumbers, roomId, userUUID, socket));
        } else {
            const sizes = ['XS', 'S', 'M', 'L', 'XL', 'XXL'];
            const sizeTooltips = ['<= 3h', '4 - 6h', '7 - 18h', '19 - 30h', '31 - 60h', '61 - 120h'];
            votingControls.appendChild(createVoteButtons(reveal, sizes, roomId, userUUID, socket, sizeTooltips));
        }
    }

    function showErrorNotification(message) {
        const notification = document.createElement('div');
        notification.classList.add('notification');
        notification.textContent = message;

        document.body.appendChild(notification);

        setTimeout(function() {
            notification.style.opacity = 1;
        }, 500);

        // Usuwa powiadomienie po 10 sekundach
        setTimeout(function() {
            notification.style.opacity = 0;
            setTimeout(function() {
                document.body.removeChild(notification);
            }, 500);
        }, 10000);
    }
});

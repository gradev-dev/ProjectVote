const CLIENT_MESSAGE_TYPE_SUMMARY = "summary";

const SERVER_MESSAGE_TYPE_SUMMARY_INFO = "summaryInfo";

document.addEventListener("DOMContentLoaded", () => {
    const container = document.querySelector(".container");
    const roomId = container.dataset.roomId;
    const body = document.querySelector("body");

    const votesHistoryContainer = container.querySelector("#votes-history");

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
        socket.send(JSON.stringify({ type: CLIENT_MESSAGE_TYPE_SUMMARY, room_id: roomId, user_name: sessionUser, session_uuid: sessionUUID, password: sessionStorage.getItem("roomPassword") }));
    };

    function getData() {
        votesHistoryContainer.innerHTML = '';



        const indicator = document.createElement('div');
        indicator.id = 'loadingHistoryIndicator';
        indicator.innerHTML = '<div class="lds-ring"><div></div><div></div><div></div><div></div></div>';

        votesHistoryContainer.appendChild(indicator);

        fetch(`/history/${roomId}?summary=true`)
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
                const votesContainerTop = document.createElement('div');
                votesContainerTop.className = 'task-top-summary';
                votesContainerTop.innerHTML = '<div class="task-top-bar"></div>';
                votesHistoryContainer.appendChild(votesContainerTop);
                data.forEach((voteData, index) => {
                    let fibonacci = voteData.fibonacci;
                    let tshirt = voteData.tshirt;


                    setTimeout(() => {
                        const taskDiv = document.createElement('div');
                        taskDiv.className = 'task-container animated-entry';

                        const headerDiv = document.createElement('div');
                        headerDiv.className = 'task-header-summary';

                        const taskTitle = document.createElement('div');
                        taskTitle.className = 'task-title';
                        taskTitle.textContent = voteData.task;

                        const taskAverage = document.createElement('div');
                        taskAverage.className = 'task-fib';
                        taskAverage.textContent = (voteData.roomMethod === 'fibonacci') ? fibonacci : tshirt;

                        const taskTime = document.createElement('div');
                        taskTime.textContent = voteData.time;


                        const taskUrlButton = document.createElement('button');
                        taskUrlButton.innerHTML = '<div><i class="bi bi-box-arrow-up-right"></i></div>';
                        taskUrlButton.onclick = function() {
                            window.open(voteData.jiraTaskUrl, '_blank');
                        };

                        headerDiv.appendChild(taskAverage);
                        headerDiv.appendChild(taskTitle);
                        headerDiv.appendChild(taskTime);
                        headerDiv.appendChild(taskUrlButton);
                        taskDiv.appendChild(headerDiv);

                        votesHistoryContainer.appendChild(taskDiv);

                    }, 100 * index);
                });
            });
    }

    getData();
    socket.onclose = (event) => {
        console.log("WebSocket closed:", event.code, event.reason);
    };
});

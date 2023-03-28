const startButton = document.getElementById('start');
const remoteVideo = document.getElementById('remoteVideo');

startButton.addEventListener('click', async () => {
    // Create a new RTCPeerConnection
    const peerConnection = new RTCPeerConnection({
        iceServers: [{ urls: 'stun:stun.l.google.com:19302' }],
    });

    // Handle the ICE candidate event
    peerConnection.onicecandidate = (event) => {
        if (event.candidate) {
            // Send the ICE candidate to the server
            fetch('/candidate', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(event.candidate),
            });
        }
    };

    // Set up the remote stream
    const remoteStream = new MediaStream();
    remoteVideo.srcObject = remoteStream;

    // Add tracks to the remote stream when they become available
    peerConnection.ontrack = (event) => {
        event.streams[0].getTracks().forEach((track) => {
            remoteStream.addTrack(track);
        });
    };

    // Fetch the offer from the server
    const response = await fetch('/watch', {
        method: 'GET',
    });

    // Receive the offer from the server and set the remote description
    const offer = await response.json();
    await peerConnection.setRemoteDescription(offer);

    // Create an answer and set the local description
    const answer = await peerConnection.createAnswer();
    await peerConnection.setLocalDescription(answer);

    // Send the answer to the server
    await fetch('/offer', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(answer),
    });
});

const startButton = document.getElementById('start');

startButton.addEventListener('click', async () => {
    // Get the media stream from the browser tab
    const stream = await navigator.mediaDevices.getDisplayMedia({
        video: { cursor: 'always' },
        audio: false,
    });

    // Create a new RTCPeerConnection
    const peerConnection = new RTCPeerConnection({
        iceServers: [{ urls: 'stun:stun.l.google.com:19302' }],
    });

    // Add the media stream tracks to the peer connection
    stream.getTracks().forEach((track) => peerConnection.addTrack(track, stream));

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

    // Create an offer and set the local description
    const offer = await peerConnection.createOffer();
    await peerConnection.setLocalDescription(offer);

    // Send the offer to the server
    const response = await fetch('/offer', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(offer),
    });

    // Receive the answer from the server and set the remote description
    const answer = await response.json();
    await peerConnection.setRemoteDescription(answer);
});

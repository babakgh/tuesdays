import asyncio
import json
import pytest
import websockets
from typing import Dict, Any

# Test configuration
WS_URL = "ws://localhost:8080/ws"
TIMEOUT = 5.0  # seconds

async def connect_client() -> websockets.WebSocketClientProtocol:
    """Helper function to create a WebSocket connection."""
    return await websockets.connect(WS_URL)

async def send_command(ws: websockets.WebSocketClientProtocol, command: str, message: str = None) -> None:
    """Helper function to send a command to the server."""
    data = {"command": command}
    if message:
        data["message"] = message
    await ws.send(json.dumps(data))

async def receive_event(ws: websockets.WebSocketClientProtocol) -> Dict[str, Any]:
    """Helper function to receive and parse an event from the server."""
    try:
        message = await asyncio.wait_for(ws.recv(), timeout=TIMEOUT)
        return json.loads(message)
    except asyncio.TimeoutError:
        pytest.fail(f"Timeout waiting for event")

@pytest.mark.asyncio
async def test_two_members_connect():
    """Test 1: Two members connect to the room"""
    # Connect two clients
    client1 = await connect_client()
    client2 = await connect_client()

    try:
        # Wait for initial messages (me command and join broadcast)
        for _ in range(3):
            await receive_event(client1)
        for _ in range(2):
            await receive_event(client2)

        # Send list command from client1
        await send_command(client1, "list")
        list_event = await receive_event(client1)

        print(list_event)
        # Verify list response
        assert list_event["event"] == "list"
        assert len(list_event["members"]) == 2
        assert all(member.startswith("member") for member in list_event["members"])

        print("✅ Test 1 passed: 2 members connected and present in storage")

    finally:
        await client1.close()
        await client2.close()

@pytest.mark.asyncio
async def test_broadcast_message():
    """Test 2: Broadcast message is delivered to all members"""
    # Connect two clients
    client1 = await connect_client()
    client2 = await connect_client()

    try:
        # Wait for initial messages
        for _ in range(3):
            await receive_event(client1)
        for _ in range(2):
            await receive_event(client2)

        # Send broadcast from client1
        test_message = "Hello from member1!"
        await send_command(client1, "broadcast", test_message)

        # Verify both clients receive the broadcast
        broadcast1 = await receive_event(client1)
        broadcast2 = await receive_event(client2)

        assert broadcast1["event"] == "broadcast"
        assert broadcast2["event"] == "broadcast"
        assert broadcast1["message"] == test_message
        assert broadcast2["message"] == test_message
        assert broadcast1["member"].startswith("member")
        assert broadcast2["member"].startswith("member")

        print("✅ Test 2 passed: Broadcast message received by all members")

    finally:
        await client1.close()
        await client2.close()

@pytest.mark.asyncio
async def test_disconnect_removes_member():
    """Test 3: Disconnect removes member from list"""
    # Connect two clients
    client1 = await connect_client()
    client2 = await connect_client()

    try:
        # Wait for initial messages
        for _ in range(3):
            await receive_event(client1)
        for _ in range(2):
            await receive_event(client2)

        # Disconnect client2
        await client2.close()

        # Send list command from client1
        await send_command(client1, "list")
        list_event = await receive_event(client1)

        # Verify only client1 is in the list
        assert list_event["event"] == "list"
        assert len(list_event["members"]) == 1
        assert list_event["members"][0].startswith("member")

        print("✅ Test 3 passed: Disconnected member removed from server list")

    finally:
        await client1.close()

@pytest.mark.asyncio
async def test_new_member_welcome():
    """Test 4: New member receives welcome commands"""
    # Connect one client
    client = await connect_client()

    try:
        # for _ in range(2):
        #     await receive_event(client)

        # Receive me command
        me_event = await receive_event(client)
        assert me_event["event"] == "me"
        assert me_event["member"].startswith("member")

        # Receive join broadcast
        join_event = await receive_event(client)
        assert join_event["event"] == "broadcast"
        assert "has joined!" in join_event["message"]

        print("✅ Test 4 passed: Member received me + joined broadcast")

    finally:
        await client.close()
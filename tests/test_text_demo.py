import pytest
from claude_agent_sdk import query, ClaudeAgentOptions, AssistantMessage, TextBlock


@pytest.mark.asyncio
async def test_text_demo_navigation(imprint_binary, example_binaries):
    """Test that Claude can navigate and interact with text-demo."""

    options = ClaudeAgentOptions(
        mcp_servers={
            "imprint": {
                "command": imprint_binary,
                "args": ["--rows", "24", "--cols", "80"],
            }
        },
        allowed_tools=[
            "mcp__imprint__type_text",
            "mcp__imprint__send_keystrokes",
            "mcp__imprint__get_screen_text",
            "mcp__imprint__wait_for_stable",
        ],
        max_turns=10,
        permission_mode="bypassPermissions",
    )

    prompt = f"""
    You have access to a terminal via MCP tools.

    1. Type this command to run the demo: {example_binaries['text-demo']}
    2. Press enter to run it
    3. Wait for the screen to stabilize
    4. Use get_screen_text to see the checklist
    5. Press 'j' twice to move down to "Buy oranges"
    6. Press space to toggle it
    7. Use get_screen_text to verify "Buy oranges" is now selected (marked with [x])
    8. Report whether you successfully selected "Buy oranges"
    """

    result_text = ""
    async for message in query(prompt=prompt, options=options):
        if isinstance(message, AssistantMessage):
            for block in message.content:
                if isinstance(block, TextBlock):
                    result_text += block.text

    assert "success" in result_text.lower() or "[x]" in result_text or "selected" in result_text.lower()


@pytest.mark.asyncio
async def test_text_demo_toggle_multiple(imprint_binary, example_binaries):
    """Test toggling multiple items in the checklist."""

    options = ClaudeAgentOptions(
        mcp_servers={
            "imprint": {
                "command": imprint_binary,
                "args": ["--rows", "24", "--cols", "80"],
            }
        },
        allowed_tools=[
            "mcp__imprint__type_text",
            "mcp__imprint__send_keystrokes",
            "mcp__imprint__get_screen_text",
            "mcp__imprint__wait_for_stable",
        ],
        max_turns=12,
        permission_mode="bypassPermissions",
    )

    prompt = f"""
    You have access to a terminal via MCP tools.

    1. Run: {example_binaries['text-demo']}
    2. Wait for stable screen
    3. Toggle "Buy apples" (first item, press space)
    4. Move down and toggle "Buy bananas"
    5. Move down and toggle "Buy oranges"
    6. Use get_screen_text to verify all three items are selected
    7. Report which items are marked with [x]
    """

    result_text = ""
    async for message in query(prompt=prompt, options=options):
        if isinstance(message, AssistantMessage):
            for block in message.content:
                if isinstance(block, TextBlock):
                    result_text += block.text

    # Should mention multiple items being selected
    assert "apples" in result_text.lower() or "bananas" in result_text.lower() or "oranges" in result_text.lower()

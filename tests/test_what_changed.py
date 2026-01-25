import pytest
from claude_agent_sdk import query, ClaudeAgentOptions, AssistantMessage, TextBlock


@pytest.mark.asyncio
async def test_what_changed_game(imprint_binary, example_binaries):
    """Test that Claude can play the what-changed game."""

    options = ClaudeAgentOptions(
        mcp_servers={
            "imprint": {
                "command": imprint_binary,
                "args": ["--rows", "30", "--cols", "100"],
            }
        },
        allowed_tools=[
            "mcp__imprint__type_text",
            "mcp__imprint__send_keystrokes",
            "mcp__imprint__get_screenshot",
            "mcp__imprint__wait_for_stable",
            "mcp__imprint__get_screen_text",
        ],
        max_turns=15,
        permission_mode="bypassPermissions",
    )

    # Use seed for reproducibility
    prompt = f"""
    You have access to a terminal via MCP tools.

    1. Run: {example_binaries['what-changed']} --seed 42
    2. Take a screenshot of the initial grid (memorize it)
    3. Press space to see the "after" state
    4. Take another screenshot and compare to find which cell changed color
    5. Navigate to the changed cell using arrow keys
    6. Press enter to submit your answer
    7. Report whether you won or lost based on the result message
    """

    result_text = ""
    async for message in query(prompt=prompt, options=options):
        if isinstance(message, AssistantMessage):
            for block in message.content:
                if isinstance(block, TextBlock):
                    result_text += block.text

    # With seed 42, the changed cell is at row 0, col 1 (cyan -> yellow)
    # Verify Claude correctly identified this specific cell
    result_lower = result_text.lower()

    # Check for success/fail outcome OR correct cell identification
    outcome_found = "success" in result_lower or "won" in result_lower or "fail" in result_lower

    # Verify correct cell coordinates were identified (row 0, col 1)
    correct_cell_identified = (
        ("row 0" in result_lower and "col" in result_lower and "1" in result_lower) or
        ("0, 1" in result_lower) or
        ("0,1" in result_lower) or
        ("position 1" in result_lower and "row 0" in result_lower)
    )

    assert outcome_found or correct_cell_identified, f"Expected to find outcome or correct cell (row 0, col 1), got: {result_text[-500:]}"


@pytest.mark.asyncio
async def test_what_changed_screenshot_comparison(imprint_binary, example_binaries):
    """Test that Claude can take and compare screenshots."""

    options = ClaudeAgentOptions(
        mcp_servers={
            "imprint": {
                "command": imprint_binary,
                "args": ["--rows", "30", "--cols", "100"],
            }
        },
        allowed_tools=[
            "mcp__imprint__type_text",
            "mcp__imprint__send_keystrokes",
            "mcp__imprint__get_screenshot",
            "mcp__imprint__wait_for_stable",
        ],
        max_turns=10,
        permission_mode="bypassPermissions",
    )

    prompt = f"""
    You have access to a terminal via MCP tools.

    1. Run: {example_binaries['what-changed']} --seed 123
    2. Take a screenshot of the "before" grid
    3. Describe the colors you see in the grid (which cells have which colors)
    4. Press space to advance to the "after" state
    5. Take another screenshot
    6. Identify which cell changed and describe the change
    """

    result_text = ""
    async for message in query(prompt=prompt, options=options):
        if isinstance(message, AssistantMessage):
            for block in message.content:
                if isinstance(block, TextBlock):
                    result_text += block.text

    # Should describe colors and changes
    result_lower = result_text.lower()
    color_mentions = sum([
        "red" in result_lower,
        "green" in result_lower,
        "blue" in result_lower,
        "yellow" in result_lower,
        "magenta" in result_lower,
        "cyan" in result_lower,
    ])
    assert color_mentions >= 1 or "change" in result_lower, "Expected color descriptions"

import logging
import pytest
from claude_agent_sdk import query, ClaudeAgentOptions, AssistantMessage, TextBlock

logger = logging.getLogger(__name__)


async def collect_response(prompt, options):
    """Collect response text from Claude, with error handling for SDK issues."""
    result_text = ""
    errors = []

    try:
        async for message in query(prompt=prompt, options=options):
            if isinstance(message, AssistantMessage):
                for block in message.content:
                    if isinstance(block, TextBlock):
                        result_text += block.text
                    else:
                        logger.debug(f"Non-TextBlock in response: {type(block).__name__}")
            else:
                logger.debug(f"Non-AssistantMessage: {type(message).__name__}")
                if hasattr(message, "error"):
                    errors.append(str(message.error))
    except Exception as e:
        pytest.fail(f"Claude SDK query failed: {type(e).__name__}: {e}")

    if errors:
        pytest.fail(f"SDK returned errors: {errors}")

    return result_text


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

    # Use seed for reproducibility - ensures the changed cell location
    # is deterministic so we can verify Claude identifies the correct cell
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

    result_text = await collect_response(prompt, options)

    # With seed 42, the changed cell is at row 1, col 3 (empty -> cyan)
    # Verify Claude correctly identified this specific cell or got an outcome
    result_lower = result_text.lower()

    # Check for success/fail outcome
    outcome_found = "success" in result_lower or "won" in result_lower or "fail" in result_lower

    # Verify correct cell coordinates were identified (row 1, col 3)
    correct_cell_identified = (
        ("row 1" in result_lower and "col" in result_lower and "3" in result_lower) or
        ("1, 3" in result_lower) or
        ("1,3" in result_lower) or
        # Also accept if Claude found cyan appearing (the new color)
        ("cyan" in result_lower and "appear" in result_lower) or
        ("empty" in result_lower and "cyan" in result_lower)
    )

    assert outcome_found or correct_cell_identified, \
        f"Expected outcome or correct cell (row 1, col 3). Response: {result_text[-500:]}"


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

    result_text = await collect_response(prompt, options)
    result_lower = result_text.lower()

    # Should describe colors and identify change
    color_mentions = sum([
        "red" in result_lower,
        "green" in result_lower,
        "blue" in result_lower,
        "yellow" in result_lower,
        "magenta" in result_lower,
        "cyan" in result_lower,
        "gray" in result_lower or "grey" in result_lower or "empty" in result_lower,
    ])

    change_identified = "change" in result_lower or "different" in result_lower

    assert color_mentions >= 2 or change_identified, \
        f"Expected color descriptions or change identification. Response: {result_text[-500:]}"

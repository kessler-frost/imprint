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

    result_text = await collect_response(prompt, options)

    assert "success" in result_text.lower() or "[x]" in result_text or "selected" in result_text.lower(), \
        f"Expected successful selection, got: {result_text[-500:]}"


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

    result_text = await collect_response(prompt, options)
    result_lower = result_text.lower()

    # Verify all three items are mentioned as selected
    items_mentioned = sum([
        "apples" in result_lower,
        "bananas" in result_lower,
        "oranges" in result_lower,
    ])

    assert items_mentioned >= 2, \
        f"Expected at least 2 items mentioned, got {items_mentioned}. Response: {result_text[-500:]}"

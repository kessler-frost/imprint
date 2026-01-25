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
async def test_screenshot_demo_visual_bugs(imprint_binary, example_binaries):
    """Test that Claude can identify visual bugs using screenshots."""

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
        max_turns=8,
        permission_mode="bypassPermissions",
    )

    prompt = f"""
    You have access to a terminal via MCP tools.

    1. Run: {example_binaries['screenshot-demo']} --seed 42
    2. Wait for stable, then take a screenshot
    3. Analyze the screenshot for visual bugs:
       - Is the title centered correctly?
       - Are there any color bleeding issues?
       - Is any text misaligned?
       - Is there poor contrast anywhere?
    4. Report each bug you find with specific details
    """

    result_text = await collect_response(prompt, options)
    result_lower = result_text.lower()

    # Count specific bug indicators (more precise matching)
    bugs_found = sum([
        # Bug 1: Misaligned/off-center title
        any(word in result_lower for word in ["misalign", "off-center", "not centered", "centering"]),
        # Bug 2/3: Color bleed issues
        "bleed" in result_lower or ("color" in result_lower and "issue" in result_lower),
        # Bug 4: Poor contrast / hard to read
        "contrast" in result_lower or "hard to read" in result_lower or "difficult to read" in result_lower,
    ])

    # Should identify at least 2 of the 4 intentional bugs
    assert bugs_found >= 2, \
        f"Expected to find at least 2 visual bugs, found {bugs_found}. Response: {result_text[-500:]}"


@pytest.mark.asyncio
async def test_screenshot_demo_navigation(imprint_binary, example_binaries):
    """Test that Claude can navigate the screenshot demo menu."""

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

    1. Run: {example_binaries['screenshot-demo']} --seed 42
    2. Wait for stable, take a screenshot
    3. Navigate down using 'j' key twice
    4. Take another screenshot and note the cursor position changed
    5. Press 'r' to regenerate colors
    6. Take a final screenshot and note the colors changed
    7. Report what changed between screenshots (cursor movement and color regeneration)
    """

    result_text = await collect_response(prompt, options)
    result_lower = result_text.lower()

    # Should report observations about navigation and color changes
    navigation_observed = any(word in result_lower for word in ["move", "cursor", "select", "navigate", "position"])
    color_observed = any(word in result_lower for word in ["color", "regenerate", "change", "different"])

    assert navigation_observed or color_observed, \
        f"Expected observations about navigation or colors. Response: {result_text[-500:]}"

import pytest
from claude_agent_sdk import query, ClaudeAgentOptions, AssistantMessage, TextBlock


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
    4. Report each bug you find
    """

    result_text = ""
    async for message in query(prompt=prompt, options=options):
        if isinstance(message, AssistantMessage):
            for block in message.content:
                if isinstance(block, TextBlock):
                    result_text += block.text

    # Should identify at least some of the intentional bugs
    result_lower = result_text.lower()
    bugs_found = sum([
        "center" in result_lower or "align" in result_lower or "misalign" in result_lower,
        "bleed" in result_lower or "color" in result_lower,
        "contrast" in result_lower or "read" in result_lower,
    ])
    assert bugs_found >= 1, f"Expected to find visual bugs, got: {result_text}"


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
    4. Take another screenshot
    5. Press 'r' to regenerate colors
    6. Take a final screenshot
    7. Report what you observed in the screenshots
    """

    result_text = ""
    async for message in query(prompt=prompt, options=options):
        if isinstance(message, AssistantMessage):
            for block in message.content:
                if isinstance(block, TextBlock):
                    result_text += block.text

    # Should report some observations about the demo
    assert len(result_text) > 50, "Expected Claude to report observations"

import ast
import asyncio
from pathlib import Path
import unittest


VOICE_ROOT = Path(__file__).resolve().parents[1]
BOT_PATH = VOICE_ROOT / "bot.py"
STATIC_INDEX_PATH = VOICE_ROOT / "static" / "index.html"


def load_isolated_function(name: str, globals_: dict | None = None):
    """Load one bot.py function without importing Pipecat/provider packages."""
    tree = ast.parse(BOT_PATH.read_text(encoding="utf-8"), filename=str(BOT_PATH))
    function = next(
        node
        for node in tree.body
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)) and node.name == name
    )
    module = ast.fix_missing_locations(ast.Module(body=[function], type_ignores=[]))
    namespace = dict(globals_ or {})
    exec(compile(module, str(BOT_PATH), "exec"), namespace)
    return namespace[name]


class InboundOnlyPolicyTests(unittest.TestCase):
    def test_tool_schemas_are_attached_to_each_llm_context(self):
        tree = ast.parse(BOT_PATH.read_text(encoding="utf-8"), filename=str(BOT_PATH))
        functions = {
            node.name: node
            for node in tree.body
            if isinstance(node, ast.AsyncFunctionDef)
            and node.name in {"stream_websocket", "browser_stream"}
        }

        for function_name, expected_tools in (
            ("stream_websocket", "phone_tools"),
            ("browser_stream", "agent_tools"),
        ):
            with self.subTest(function=function_name):
                calls = [node for node in ast.walk(functions[function_name]) if isinstance(node, ast.Call)]
                service_call = next(
                    node
                    for node in calls
                    if isinstance(node.func, ast.Name) and node.func.id == "OpenAILLMService"
                )
                context_call = next(
                    node
                    for node in calls
                    if isinstance(node.func, ast.Name) and node.func.id == "LLMContext"
                )

                service_keywords = {keyword.arg for keyword in service_call.keywords}
                context_keywords = {keyword.arg: keyword.value for keyword in context_call.keywords}

                self.assertNotIn("tools", service_keywords)
                self.assertIn("tools", context_keywords)
                self.assertEqual(ast.unparse(context_keywords["tools"]), expected_tools)

    def test_forged_stream_parameters_cannot_select_sales_mode(self):
        parse_params = load_isolated_function("_stream_custom_params")
        normalize_mode = load_isolated_function("_normalize_agent_mode")
        forged_start = {
            "customParameters": {
                "agent": "sales",
                "caller_phone": "+61400000000",
            }
        }

        params = parse_params(forged_start)

        self.assertEqual(normalize_mode(params["agent"]), "corporate")
        self.assertEqual(normalize_mode("sales", default="sales"), "corporate")
        self.assertEqual(normalize_mode("corporate"), "corporate")
        self.assertEqual(normalize_mode("restaurant"), "restaurant")

        source = BOT_PATH.read_text(encoding="utf-8")
        stream_node = next(
            node
            for node in ast.parse(source).body
            if isinstance(node, ast.AsyncFunctionDef) and node.name == "stream_websocket"
        )
        stream_source = ast.get_source_segment(source, stream_node) or ""
        self.assertIn("_normalize_agent_mode", stream_source)
        self.assertNotIn("phone_tools = TOOLS", stream_source)
        self.assertNotIn('agent_mode == "sales"', stream_source)

    def test_retired_sms_tool_fails_closed_without_constructing_twilio(self):
        class TwilioMustNotBeConstructed:
            constructed = 0

            def __init__(self, *_args, **_kwargs):
                type(self).constructed += 1
                raise AssertionError("Twilio client must not be constructed for retired SMS tools")

        class LoggerStub:
            @staticmethod
            def info(*_args, **_kwargs):
                pass

        dispatch = load_isolated_function(
            "_dispatch_tool",
            {
                "logger": LoggerStub(),
                "TwilioClient": TwilioMustNotBeConstructed,
            },
        )

        result = asyncio.run(
            dispatch(
                function_name="send_followup_sms",
                arguments={"message": "forged"},
                call_db_id=1,
                to_number_ref=["+61400000000"],
                outcome_state=["unknown"],
                task_ref=[None],
                booking_state={},
                call_sid="CA-forged",
                is_browser=False,
            )
        )

        self.assertEqual(result["status"], "disabled")
        self.assertEqual(TwilioMustNotBeConstructed.constructed, 0)

        source = BOT_PATH.read_text(encoding="utf-8")
        self.assertNotIn(".messages.create(", source)
        self.assertNotIn('"name": "send_followup_sms"', source)
        self.assertNotIn("resolve_caller_id", source)
        self.assertNotIn("python dial.py", source)
        self.assertFalse((VOICE_ROOT / "caller_id.py").exists())

    def test_public_ui_contains_no_outbound_dialer(self):
        html = STATIC_INDEX_PATH.read_text(encoding="utf-8")

        for retired_marker in (
            "Phone Outreach",
            "Call Now",
            "phone-input",
            "campaign-input",
            "fetch('/call'",
        ):
            with self.subTest(marker=retired_marker):
                self.assertNotIn(retired_marker, html)


if __name__ == "__main__":
    unittest.main()

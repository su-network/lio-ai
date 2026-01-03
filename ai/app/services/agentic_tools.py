"""
Agentic Tool Definitions for LLM Function Calling

This module provides a comprehensive set of tool/function definitions that
AI models can use to perform various tasks autonomously.
"""

from typing import List, Dict, Any
import subprocess
import os
import json
import logging

logger = logging.getLogger(__name__)


# Tool definitions in OpenAI function calling format
AVAILABLE_TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "execute_python_code",
            "description": "Execute Python code in a sandboxed environment and return the output. Use this for calculations, data processing, or generating code examples.",
            "parameters": {
                "type": "object",
                "properties": {
                    "code": {
                        "type": "string",
                        "description": "The Python code to execute"
                    },
                    "timeout": {
                        "type": "integer",
                        "description": "Maximum execution time in seconds (default: 10)",
                        "default": 10
                    }
                },
                "required": ["code"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "search_web",
            "description": "Search the web for current information. Use this when you need up-to-date information or facts that you weren't trained on.",
            "parameters": {
                "type": "object",
                "properties": {
                    "query": {
                        "type": "string",
                        "description": "The search query"
                    },
                    "num_results": {
                        "type": "integer",
                        "description": "Number of results to return (default: 5)",
                        "default": 5
                    }
                },
                "required": ["query"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "read_file",
            "description": "Read the contents of a file from the filesystem. Use this to access configuration files, code files, or data files.",
            "parameters": {
                "type": "object",
                "properties": {
                    "file_path": {
                        "type": "string",
                        "description": "The path to the file to read"
                    },
                    "encoding": {
                        "type": "string",
                        "description": "File encoding (default: utf-8)",
                        "default": "utf-8"
                    }
                },
                "required": ["file_path"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "write_file",
            "description": "Write content to a file on the filesystem. Use this to save generated code, data, or configurations.",
            "parameters": {
                "type": "object",
                "properties": {
                    "file_path": {
                        "type": "string",
                        "description": "The path where the file should be written"
                    },
                    "content": {
                        "type": "string",
                        "description": "The content to write to the file"
                    },
                    "encoding": {
                        "type": "string",
                        "description": "File encoding (default: utf-8)",
                        "default": "utf-8"
                    }
                },
                "required": ["file_path", "content"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "list_directory",
            "description": "List files and directories in a given path. Use this to explore the filesystem structure.",
            "parameters": {
                "type": "object",
                "properties": {
                    "directory_path": {
                        "type": "string",
                        "description": "The directory path to list"
                    },
                    "recursive": {
                        "type": "boolean",
                        "description": "Whether to list recursively (default: false)",
                        "default": False
                    }
                },
                "required": ["directory_path"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "execute_shell_command",
            "description": "Execute a shell command and return its output. Use this for system operations, running tests, or interacting with CLI tools.",
            "parameters": {
                "type": "object",
                "properties": {
                    "command": {
                        "type": "string",
                        "description": "The shell command to execute"
                    },
                    "working_directory": {
                        "type": "string",
                        "description": "The working directory for the command (optional)"
                    },
                    "timeout": {
                        "type": "integer",
                        "description": "Maximum execution time in seconds (default: 30)",
                        "default": 30
                    }
                },
                "required": ["command"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "calculate",
            "description": "Perform mathematical calculations. Use this for complex math operations, statistics, or numerical analysis.",
            "parameters": {
                "type": "object",
                "properties": {
                    "expression": {
                        "type": "string",
                        "description": "The mathematical expression to evaluate (Python syntax)"
                    }
                },
                "required": ["expression"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "analyze_code",
            "description": "Analyze code for potential issues, security vulnerabilities, or improvements. Use this for code review tasks.",
            "parameters": {
                "type": "object",
                "properties": {
                    "code": {
                        "type": "string",
                        "description": "The code to analyze"
                    },
                    "language": {
                        "type": "string",
                        "description": "Programming language (python, javascript, go, etc.)"
                    },
                    "analysis_type": {
                        "type": "string",
                        "description": "Type of analysis: security, performance, style, or all",
                        "enum": ["security", "performance", "style", "all"],
                        "default": "all"
                    }
                },
                "required": ["code", "language"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "generate_documentation",
            "description": "Generate documentation for code. Use this to create README files, API docs, or code comments.",
            "parameters": {
                "type": "object",
                "properties": {
                    "code": {
                        "type": "string",
                        "description": "The code to document"
                    },
                    "language": {
                        "type": "string",
                        "description": "Programming language"
                    },
                    "doc_format": {
                        "type": "string",
                        "description": "Documentation format: markdown, html, or docstring",
                        "enum": ["markdown", "html", "docstring"],
                        "default": "markdown"
                    }
                },
                "required": ["code", "language"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "run_tests",
            "description": "Run tests for a given code snippet or file. Use this to verify code correctness.",
            "parameters": {
                "type": "object",
                "properties": {
                    "test_file": {
                        "type": "string",
                        "description": "Path to the test file"
                    },
                    "test_framework": {
                        "type": "string",
                        "description": "Testing framework: pytest, unittest, jest, go test, etc."
                    }
                },
                "required": ["test_file", "test_framework"]
            }
        }
    }
]


# Tool execution functions
async def execute_tool(tool_name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
    """
    Execute a tool function and return the result
    
    Args:
        tool_name: Name of the tool to execute
        arguments: Arguments for the tool
        
    Returns:
        Dict with result or error
    """
    try:
        if tool_name == "execute_python_code":
            return await execute_python_code(**arguments)
        elif tool_name == "search_web":
            return await search_web(**arguments)
        elif tool_name == "read_file":
            return await read_file(**arguments)
        elif tool_name == "write_file":
            return await write_file(**arguments)
        elif tool_name == "list_directory":
            return await list_directory(**arguments)
        elif tool_name == "execute_shell_command":
            return await execute_shell_command(**arguments)
        elif tool_name == "calculate":
            return await calculate(**arguments)
        elif tool_name == "analyze_code":
            return await analyze_code(**arguments)
        elif tool_name == "generate_documentation":
            return await generate_documentation(**arguments)
        elif tool_name == "run_tests":
            return await run_tests(**arguments)
        else:
            return {
                "success": False,
                "error": f"Unknown tool: {tool_name}"
            }
    except Exception as e:
        logger.error(f"Tool execution failed: {tool_name}, error: {str(e)}")
        return {
            "success": False,
            "error": str(e)
        }


async def execute_python_code(code: str, timeout: int = 10) -> Dict[str, Any]:
    """Execute Python code in a subprocess"""
    try:
        result = subprocess.run(
            ["python3", "-c", code],
            capture_output=True,
            text=True,
            timeout=timeout
        )
        
        return {
            "success": result.returncode == 0,
            "output": result.stdout,
            "error": result.stderr if result.returncode != 0 else None
        }
    except subprocess.TimeoutExpired:
        return {
            "success": False,
            "error": f"Code execution timed out after {timeout} seconds"
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


async def search_web(query: str, num_results: int = 5) -> Dict[str, Any]:
    """Search the web (placeholder - would integrate with search API)"""
    return {
        "success": True,
        "message": "Web search is not yet implemented. This is a placeholder.",
        "query": query,
        "results": []
    }


async def read_file(file_path: str, encoding: str = "utf-8") -> Dict[str, Any]:
    """Read file contents"""
    try:
        # Security: Only allow reading from specific directories
        allowed_dirs = ["/workspaces", "/tmp", "./"]
        
        if not any(file_path.startswith(d) for d in allowed_dirs):
            return {
                "success": False,
                "error": "Access to this directory is not allowed"
            }
        
        with open(file_path, "r", encoding=encoding) as f:
            content = f.read()
        
        return {
            "success": True,
            "content": content,
            "file_path": file_path
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


async def write_file(file_path: str, content: str, encoding: str = "utf-8") -> Dict[str, Any]:
    """Write content to file"""
    try:
        # Security: Only allow writing to specific directories
        allowed_dirs = ["/workspaces", "/tmp", "./"]
        
        if not any(file_path.startswith(d) for d in allowed_dirs):
            return {
                "success": False,
                "error": "Access to this directory is not allowed"
            }
        
        # Create directory if it doesn't exist
        os.makedirs(os.path.dirname(file_path), exist_ok=True)
        
        with open(file_path, "w", encoding=encoding) as f:
            f.write(content)
        
        return {
            "success": True,
            "message": f"File written successfully: {file_path}",
            "file_path": file_path,
            "bytes_written": len(content.encode(encoding))
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


async def list_directory(directory_path: str, recursive: bool = False) -> Dict[str, Any]:
    """List directory contents"""
    try:
        if recursive:
            items = []
            for root, dirs, files in os.walk(directory_path):
                for name in dirs + files:
                    items.append(os.path.join(root, name))
        else:
            items = [os.path.join(directory_path, item) for item in os.listdir(directory_path)]
        
        return {
            "success": True,
            "items": items,
            "count": len(items)
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


async def execute_shell_command(
    command: str,
    working_directory: str = None,
    timeout: int = 30
) -> Dict[str, Any]:
    """Execute shell command"""
    try:
        result = subprocess.run(
            command,
            shell=True,
            capture_output=True,
            text=True,
            cwd=working_directory,
            timeout=timeout
        )
        
        return {
            "success": result.returncode == 0,
            "output": result.stdout,
            "error": result.stderr if result.returncode != 0 else None,
            "return_code": result.returncode
        }
    except subprocess.TimeoutExpired:
        return {
            "success": False,
            "error": f"Command timed out after {timeout} seconds"
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


async def calculate(expression: str) -> Dict[str, Any]:
    """Evaluate mathematical expression"""
    try:
        # Use eval with restricted namespace for safety
        allowed_names = {
            "abs": abs, "round": round, "min": min, "max": max,
            "sum": sum, "pow": pow, "len": len,
        }
        
        # Import math functions
        import math
        allowed_names.update({
            name: getattr(math, name)
            for name in dir(math) if not name.startswith("_")
        })
        
        result = eval(expression, {"__builtins__": {}}, allowed_names)
        
        return {
            "success": True,
            "result": result,
            "expression": expression
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


async def analyze_code(
    code: str,
    language: str,
    analysis_type: str = "all"
) -> Dict[str, Any]:
    """Analyze code (basic implementation)"""
    # This is a simplified implementation
    # In production, integrate with tools like pylint, flake8, bandit, etc.
    
    issues = []
    
    # Basic checks
    if language == "python":
        if "eval(" in code:
            issues.append({
                "type": "security",
                "severity": "high",
                "message": "Use of eval() can be dangerous"
            })
        if "exec(" in code:
            issues.append({
                "type": "security",
                "severity": "high",
                "message": "Use of exec() can be dangerous"
            })
        if "import os" in code and "os.system" in code:
            issues.append({
                "type": "security",
                "severity": "medium",
                "message": "Use of os.system() can be dangerous"
            })
    
    return {
        "success": True,
        "language": language,
        "analysis_type": analysis_type,
        "issues": issues,
        "issue_count": len(issues)
    }


async def generate_documentation(
    code: str,
    language: str,
    doc_format: str = "markdown"
) -> Dict[str, Any]:
    """Generate documentation (placeholder)"""
    return {
        "success": True,
        "message": "Documentation generation is a placeholder. Use an LLM to generate actual docs.",
        "language": language,
        "format": doc_format
    }


async def run_tests(test_file: str, test_framework: str) -> Dict[str, Any]:
    """Run tests"""
    try:
        if test_framework == "pytest":
            command = f"pytest {test_file} -v"
        elif test_framework == "unittest":
            command = f"python -m unittest {test_file}"
        elif test_framework == "jest":
            command = f"jest {test_file}"
        elif test_framework == "go test":
            command = f"go test {test_file}"
        else:
            return {
                "success": False,
                "error": f"Unsupported test framework: {test_framework}"
            }
        
        result = subprocess.run(
            command,
            shell=True,
            capture_output=True,
            text=True,
            timeout=60
        )
        
        return {
            "success": result.returncode == 0,
            "output": result.stdout,
            "error": result.stderr if result.returncode != 0 else None,
            "return_code": result.returncode
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


def get_tools_for_task(task_type: str = "general") -> List[Dict[str, Any]]:
    """
    Get relevant tools for a specific task type
    
    Args:
        task_type: Type of task (coding, analysis, documentation, general)
        
    Returns:
        List of relevant tool definitions
    """
    if task_type == "coding":
        return [t for t in AVAILABLE_TOOLS if t["function"]["name"] in [
            "execute_python_code", "read_file", "write_file", "list_directory",
            "execute_shell_command", "analyze_code", "run_tests"
        ]]
    elif task_type == "analysis":
        return [t for t in AVAILABLE_TOOLS if t["function"]["name"] in [
            "analyze_code", "calculate", "read_file", "execute_python_code"
        ]]
    elif task_type == "documentation":
        return [t for t in AVAILABLE_TOOLS if t["function"]["name"] in [
            "generate_documentation", "read_file", "write_file", "analyze_code"
        ]]
    else:
        # Return all tools for general tasks
        return AVAILABLE_TOOLS

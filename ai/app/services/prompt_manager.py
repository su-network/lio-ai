"""Prompt management service for loading and formatting prompts"""
from typing import Dict, Any
import yaml
import logging
from pathlib import Path

logger = logging.getLogger(__name__)


class PromptManager:
    """Manager for loading and formatting code generation prompts"""
    
    def __init__(self, config_path: str):
        self.config_path = config_path
        self.system_prompts: Dict[str, str] = {}
        self.language_templates: Dict[str, Dict[str, str]] = {}
        self.framework_instructions: Dict[str, Dict[str, str]] = {}
        self.test_templates: Dict[str, str] = {}
        self.documentation_templates: Dict[str, str] = {}
        self.constraint_templates: Dict[str, str] = {}
        self._load_config()
    
    def _load_config(self):
        """Load prompts configuration from YAML"""
        try:
            with open(self.config_path, 'r') as f:
                config = yaml.safe_load(f)
            
            self.system_prompts = config.get('system_prompts', {})
            self.language_templates = config.get('language_templates', {})
            self.framework_instructions = config.get('framework_instructions', {})
            self.test_templates = config.get('test_templates', {})
            self.documentation_templates = config.get('documentation_templates', {})
            self.constraint_templates = config.get('constraint_templates', {})
            
            logger.info(f"Loaded prompts for {len(self.language_templates)} languages")
            
        except Exception as e:
            logger.error(f"Failed to load prompts config: {str(e)}")
            raise
    
    def build_prompt(
        self,
        language: str,
        complexity: str,
        prompt: str,
        framework: str = None,
        context: str = None,
        constraints: list = None,
        include_tests: bool = False,
        include_docs: bool = True
    ) -> str:
        """Build complete prompt from template"""
        
        # Get language template
        lang_templates = self.language_templates.get(language, {})
        template = lang_templates.get(complexity, lang_templates.get("intermediate", ""))
        
        if not template:
            # Fallback to generic template
            template = self._get_generic_template(complexity)
        
        # Get framework instruction
        framework_instruction = ""
        if framework:
            fw_config = self.framework_instructions.get(framework, {})
            framework_instruction = fw_config.get("instruction", "")
        
        # Get constraint instructions
        constraints_instruction = ""
        if constraints:
            constraint_texts = []
            for constraint in constraints:
                if constraint in self.constraint_templates:
                    constraint_texts.append(self.constraint_templates[constraint])
            if constraint_texts:
                constraints_instruction = "Additional Constraints:\n" + "\n".join(
                    f"- {c}" for c in constraint_texts
                )
        
        # Format template
        formatted_prompt = template.format(
            prompt=prompt,
            framework_instruction=framework_instruction,
            constraints_instruction=constraints_instruction,
            context=context or "No additional context provided."
        )
        
        # Add system prompt
        system_prompt = self.system_prompts.get("code_generation", "")
        full_prompt = f"{system_prompt}\n\n{formatted_prompt}"
        
        # Add test generation instruction
        if include_tests:
            full_prompt += "\n\nAlso generate comprehensive unit tests for the code."
        
        # Add documentation instruction
        if include_docs:
            full_prompt += "\n\nInclude detailed documentation and usage examples."
        
        return full_prompt
    
    def _get_generic_template(self, complexity: str) -> str:
        """Get generic template for unsupported language"""
        templates = {
            "simple": """
Generate code for: {prompt}

Requirements:
- Clean, readable code
- Basic error handling
- Comments for key logic
{framework_instruction}
{constraints_instruction}

Context: {context}
""",
            "intermediate": """
Generate production-ready code for: {prompt}

Requirements:
- Best practices and idioms
- Comprehensive error handling
- Detailed documentation
- Input validation
{framework_instruction}
{constraints_instruction}

Context: {context}
""",
            "advanced": """
Generate enterprise-grade code for: {prompt}

Requirements:
- Advanced design patterns
- Robust error handling
- Performance optimization
- Security considerations
- Comprehensive testing
- Full documentation
{framework_instruction}
{constraints_instruction}

Context: {context}
"""
        }
        return templates.get(complexity, templates["intermediate"])
    
    def build_test_prompt(self, code: str, language: str) -> str:
        """Build prompt for test generation"""
        template = self.test_templates.get(
            language,
            "Generate comprehensive unit tests for the following code:\n\n{code}"
        )
        return template.format(code=code)
    
    def build_documentation_prompt(self, code: str, doc_type: str = "readme") -> str:
        """Build prompt for documentation generation"""
        template = self.documentation_templates.get(
            doc_type,
            "Generate comprehensive documentation for:\n\n{code}"
        )
        return template.format(code=code)
    
    def get_system_prompt(self, prompt_type: str = "code_generation") -> str:
        """Get system prompt by type"""
        return self.system_prompts.get(prompt_type, self.system_prompts.get("base", ""))
    
    def reload_config(self):
        """Reload configuration from file"""
        self._load_config()
        logger.info("Prompts configuration reloaded")

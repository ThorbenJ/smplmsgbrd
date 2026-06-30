# Contributing

Thank you for your interest in contributing. This document covers how to submit
changes, licensing expectations, and the project's policy on AI-assisted development.

## Getting started

- For setup and build instructions, see [README.md](README.md).
- If a `CODING_STYLE.md` exists, read it before writing any code.
- There is no formal issue tracker requirement for small fixes, but opening an
  issue before large changes avoids wasted effort.

## Submitting changes

1. Fork the repository and work on a dedicated branch.
2. Keep commits focused — one logical change per commit.
3. Open a pull request with a clear description of what changed and why.

## Coding style

Follow the conventions of the language you are working in. If a `CODING_STYLE.md`
exists in the repository, it takes precedence.

Do not add unnecessary comments. Explain *why* only when it is genuinely
non-obvious — not what the code does.

## Contributor licence agreement

By opening a pull request you agree to the following:

> You certify that your contribution is your original work (or that you have
> the right to submit it), and you irrevocably assign and transfer all
> copyright and other intellectual property rights in your contribution to the
> project owners. The project owners reserve the right to relicence the
> project, including all contributions, under any licence they choose at any
> time.

This is a lightweight inbound=outbound assignment. It allows the project to
remain coherent under a single copyright holder and to adapt its licence as
the project evolves, without requiring contributors to sign a separate
document. If you are not able to agree to these terms, please do not submit a
pull request.

See [LICENSE](LICENSE) for the current licence.

## AI assistance policy

AI-assisted contributions are welcome. The following rules apply to any
contribution where generative AI tools were used at any point.

**1. Do not misrepresent AI use.**  
Do not claim personal authorship of AI-generated code. You are not required to
enumerate which tools you used, but you must not actively obscure or deny it.

**2. Understand everything you submit.**  
You must be able to explain every change in your pull request — its purpose,
its logic, and its side-effects. Reviewers may ask you to walk through any
part of the diff. If you cannot answer those questions, the contribution is
not ready. This is the core requirement: a PR summary and a passing diff are
not a substitute for genuine understanding of the code.

**3. Review all changes before opening the PR.**  
All changes must be read and understood by the submitting person before the
pull request is opened. Agentic or automated workflows that apply edits
without a person reviewing each change significantly increase the risk of
submitting code the author does not truly understand — use them with extra
caution and apply heightened scrutiny before submitting.

**4. You are accountable.**  
The submitter is solely responsible for all code in a pull request, regardless
of how it was generated. AI tools are development aids, not co-authors. This
responsibility cannot be delegated.

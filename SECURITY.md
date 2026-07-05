# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| 0.2.x   | ✅        |

## Reporting

Reporte vulnerabilidades **privadamente** (não abra issue pública com exploit).

- GitHub: [Security Advisories](https://github.com/sraphaz/arah-harness/security/advisories/new)
- Ou contacto via maintainer do repositório

## Scope

Este projeto distribui **manifests, scripts e hooks** para desenvolvimento assistido por agentes. Não executa código de produto nem expõe serviços de rede.

Riscos a considerar ao instalar em seu repo:

- Scripts PowerShell com acesso ao filesystem do projeto
- Hooks Cursor que rodam ao final de turnos do agente
- Workflows GitHub Actions copiados para `.github/workflows/`

Revise diffs após `arah init` e antes de merge, especialmente em repositórios com secrets.

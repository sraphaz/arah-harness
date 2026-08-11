# Branch protection e hooks locais

Mitiga o risco “enforcement só por cooperação” (análise técnica §5).

## Local — pre-commit

```powershell
powershell -File cli/arah.ps1 hooks install -Target .
```

O hook (`templates/git-hooks/pre-commit`):

1. Recusa stage de `.arah/local/` (estado quente).
2. Falha em padrões óbvios de secret no diff staged.
3. Roda `validate-manifests.ps1` quando PowerShell estiver disponível.

## Remoto — GitHub

Recomendações mínimas em Settings → Branches → protection rule para `main`:

- [x] Require a pull request before merging  
- [x] Require status checks (`agents-validate` / self-test)  
- [x] Require conversation resolution  
- [ ] Restrict who can push (opcional)  
- [x] Do not allow bypassing the above settings (admins inclusive, se possível)

## CI

Workflow `agents-validate` / `self-test` continua sendo a fonte de verdade de gates. Hooks locais são **acelerador**, não substituto.

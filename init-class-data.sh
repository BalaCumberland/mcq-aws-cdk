#!/bin/bash

echo "🚀 Initializing class data..."

API_URL="https://ieetpwfoci.execute-api.us-east-1.amazonaws.com/prod"

# Insert classes
CLASSES=("CLS6" "CLS7" "CLS8" "CLS9" "CLS10" "CLS11-MPC" "CLS12-MPC" "CLS11-BIPC" "CLS12-BIPC")

for class in "${CLASSES[@]}"; do
    curl -X POST "$API_URL/v2/class/insert" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjZkZTQwZjA0ODgxYzZhMDE2MTFlYjI4NGE0Yzk1YTI1MWU5MTEyNTAiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vZ290aGlua2Vyc3RlY2giLCJhdWQiOiJnb3RoaW5rZXJzdGVjaCIsImF1dGhfdGltZSI6MTc1MzMzNDU0NCwidXNlcl9pZCI6IlBkbVR6dkxROVFVdXVPUllyQmg4a1dEZTFmeDEiLCJzdWIiOiJQZG1UenZMUTlRVXV1T1JZckJoOGtXRGUxZngxIiwiaWF0IjoxNzUzNDE5NDcwLCJleHAiOjE3NTM0MjMwNzAsImVtYWlsIjoicmd2dmFybWEwMDlAZ21haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImZpcmViYXNlIjp7ImlkZW50aXRpZXMiOnsiZW1haWwiOlsicmd2dmFybWEwMDlAZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoicGFzc3dvcmQifX0.M-2cd1C9TaeX77COs7uFuaODXGX9yAbu3OPpsa-4L_xB87SAiLDRVGh6vEJTWVoe10OiZgtmUJbU0Cb6tTS-CSGICMJURHR3jEkcGcpWeF418Xb0Sctr3ldLSK_u86aZtuYawobhKRKPct_TksPpDkY9EKwAi0VsC8jNrY8okUSuqx8nHEUwnDxRIYJErxmsS6hN91LUnGEsZyoHG_LMUoyB3bdyxuGSJEkn7PDg3CFLUvmyfyMchzm5T2mHoPZMTQ47Dn38CDFy4mb5GXkWAZ7JqwLxK8MlBDpkt2e2ykePIJ_dhE3qnxgfhC-TFDCqhMXdGxayd6hCSESGi9bdGA" \
        -d "{\"className\":\"$class\"}"
    echo "✅ Inserted class: $class"
done

# Insert subjects for CLS6
for subject in TELUGU HINDI ENGLISH MATHS SCIENCE SOCIAL; do
    curl -X POST "$API_URL/v2/subject/insert" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjZkZTQwZjA0ODgxYzZhMDE2MTFlYjI4NGE0Yzk1YTI1MWU5MTEyNTAiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vZ290aGlua2Vyc3RlY2giLCJhdWQiOiJnb3RoaW5rZXJzdGVjaCIsImF1dGhfdGltZSI6MTc1MzMzNDU0NCwidXNlcl9pZCI6IlBkbVR6dkxROVFVdXVPUllyQmg4a1dEZTFmeDEiLCJzdWIiOiJQZG1UenZMUTlRVXV1T1JZckJoOGtXRGUxZngxIiwiaWF0IjoxNzUzNDE5NDcwLCJleHAiOjE3NTM0MjMwNzAsImVtYWlsIjoicmd2dmFybWEwMDlAZ21haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImZpcmViYXNlIjp7ImlkZW50aXRpZXMiOnsiZW1haWwiOlsicmd2dmFybWEwMDlAZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoicGFzc3dvcmQifX0.M-2cd1C9TaeX77COs7uFuaODXGX9yAbu3OPpsa-4L_xB87SAiLDRVGh6vEJTWVoe10OiZgtmUJbU0Cb6tTS-CSGICMJURHR3jEkcGcpWeF418Xb0Sctr3ldLSK_u86aZtuYawobhKRKPct_TksPpDkY9EKwAi0VsC8jNrY8okUSuqx8nHEUwnDxRIYJErxmsS6hN91LUnGEsZyoHG_LMUoyB3bdyxuGSJEkn7PDg3CFLUvmyfyMchzm5T2mHoPZMTQ47Dn38CDFy4mb5GXkWAZ7JqwLxK8MlBDpkt2e2ykePIJ_dhE3qnxgfhC-TFDCqhMXdGxayd6hCSESGi9bdGA" \
        -d "{\"className\":\"CLS6\",\"subjectName\":\"$subject\"}"
    echo "✅ Inserted subject: CLS6 -> $subject"
done

# Insert subjects for CLS7
for subject in TELUGU HINDI ENGLISH MATHS SCIENCE SOCIAL; do
    curl -X POST "$API_URL/v2/subject/insert" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjZkZTQwZjA0ODgxYzZhMDE2MTFlYjI4NGE0Yzk1YTI1MWU5MTEyNTAiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vZ290aGlua2Vyc3RlY2giLCJhdWQiOiJnb3RoaW5rZXJzdGVjaCIsImF1dGhfdGltZSI6MTc1MzMzNDU0NCwidXNlcl9pZCI6IlBkbVR6dkxROVFVdXVPUllyQmg4a1dEZTFmeDEiLCJzdWIiOiJQZG1UenZMUTlRVXV1T1JZckJoOGtXRGUxZngxIiwiaWF0IjoxNzUzNDE5NDcwLCJleHAiOjE3NTM0MjMwNzAsImVtYWlsIjoicmd2dmFybWEwMDlAZ21haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImZpcmViYXNlIjp7ImlkZW50aXRpZXMiOnsiZW1haWwiOlsicmd2dmFybWEwMDlAZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoicGFzc3dvcmQifX0.M-2cd1C9TaeX77COs7uFuaODXGX9yAbu3OPpsa-4L_xB87SAiLDRVGh6vEJTWVoe10OiZgtmUJbU0Cb6tTS-CSGICMJURHR3jEkcGcpWeF418Xb0Sctr3ldLSK_u86aZtuYawobhKRKPct_TksPpDkY9EKwAi0VsC8jNrY8okUSuqx8nHEUwnDxRIYJErxmsS6hN91LUnGEsZyoHG_LMUoyB3bdyxuGSJEkn7PDg3CFLUvmyfyMchzm5T2mHoPZMTQ47Dn38CDFy4mb5GXkWAZ7JqwLxK8MlBDpkt2e2ykePIJ_dhE3qnxgfhC-TFDCqhMXdGxayd6hCSESGi9bdGA" \
        -d "{\"className\":\"CLS7\",\"subjectName\":\"$subject\"}"
    echo "✅ Inserted subject: CLS7 -> $subject"
done

# Insert subjects for CLS8
for subject in TELUGU HINDI ENGLISH MATHS SCIENCE SOCIAL; do
    curl -X POST "$API_URL/v2/subject/insert" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjZkZTQwZjA0ODgxYzZhMDE2MTFlYjI4NGE0Yzk1YTI1MWU5MTEyNTAiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vZ290aGlua2Vyc3RlY2giLCJhdWQiOiJnb3RoaW5rZXJzdGVjaCIsImF1dGhfdGltZSI6MTc1MzMzNDU0NCwidXNlcl9pZCI6IlBkbVR6dkxROVFVdXVPUllyQmg4a1dEZTFmeDEiLCJzdWIiOiJQZG1UenZMUTlRVXV1T1JZckJoOGtXRGUxZngxIiwiaWF0IjoxNzUzNDE5NDcwLCJleHAiOjE3NTM0MjMwNzAsImVtYWlsIjoicmd2dmFybWEwMDlAZ21haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImZpcmViYXNlIjp7ImlkZW50aXRpZXMiOnsiZW1haWwiOlsicmd2dmFybWEwMDlAZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoicGFzc3dvcmQifX0.M-2cd1C9TaeX77COs7uFuaODXGX9yAbu3OPpsa-4L_xB87SAiLDRVGh6vEJTWVoe10OiZgtmUJbU0Cb6tTS-CSGICMJURHR3jEkcGcpWeF418Xb0Sctr3ldLSK_u86aZtuYawobhKRKPct_TksPpDkY9EKwAi0VsC8jNrY8okUSuqx8nHEUwnDxRIYJErxmsS6hN91LUnGEsZyoHG_LMUoyB3bdyxuGSJEkn7PDg3CFLUvmyfyMchzm5T2mHoPZMTQ47Dn38CDFy4mb5GXkWAZ7JqwLxK8MlBDpkt2e2ykePIJ_dhE3qnxgfhC-TFDCqhMXdGxayd6hCSESGi9bdGA" \
        -d "{\"className\":\"CLS8\",\"subjectName\":\"$subject\"}"
    echo "✅ Inserted subject: CLS8 -> $subject"
done

# Insert subjects for CLS9
for subject in TELUGU HINDI ENGLISH MATHS SCIENCE SOCIAL; do
    curl -X POST "$API_URL/v2/subject/insert" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjZkZTQwZjA0ODgxYzZhMDE2MTFlYjI4NGE0Yzk1YTI1MWU5MTEyNTAiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vZ290aGlua2Vyc3RlY2giLCJhdWQiOiJnb3RoaW5rZXJzdGVjaCIsImF1dGhfdGltZSI6MTc1MzMzNDU0NCwidXNlcl9pZCI6IlBkbVR6dkxROVFVdXVPUllyQmg4a1dEZTFmeDEiLCJzdWIiOiJQZG1UenZMUTlRVXV1T1JZckJoOGtXRGUxZngxIiwiaWF0IjoxNzUzNDE5NDcwLCJleHAiOjE3NTM0MjMwNzAsImVtYWlsIjoicmd2dmFybWEwMDlAZ21haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImZpcmViYXNlIjp7ImlkZW50aXRpZXMiOnsiZW1haWwiOlsicmd2dmFybWEwMDlAZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoicGFzc3dvcmQifX0.M-2cd1C9TaeX77COs7uFuaODXGX9yAbu3OPpsa-4L_xB87SAiLDRVGh6vEJTWVoe10OiZgtmUJbU0Cb6tTS-CSGICMJURHR3jEkcGcpWeF418Xb0Sctr3ldLSK_u86aZtuYawobhKRKPct_TksPpDkY9EKwAi0VsC8jNrY8okUSuqx8nHEUwnDxRIYJErxmsS6hN91LUnGEsZyoHG_LMUoyB3bdyxuGSJEkn7PDg3CFLUvmyfyMchzm5T2mHoPZMTQ47Dn38CDFy4mb5GXkWAZ7JqwLxK8MlBDpkt2e2ykePIJ_dhE3qnxgfhC-TFDCqhMXdGxayd6hCSESGi9bdGA" \
        -d "{\"className\":\"CLS9\",\"subjectName\":\"$subject\"}"
    echo "✅ Inserted subject: CLS9 -> $subject"
done

# Insert subjects for CLS10
for subject in TELUGU HINDI ENGLISH MATHS SCIENCE SOCIAL BRIDGE POLYTECHNIC FORMULAS; do
    curl -X POST "$API_URL/v2/subject/insert" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjZkZTQwZjA0ODgxYzZhMDE2MTFlYjI4NGE0Yzk1YTI1MWU5MTEyNTAiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vZ290aGlua2Vyc3RlY2giLCJhdWQiOiJnb3RoaW5rZXJzdGVjaCIsImF1dGhfdGltZSI6MTc1MzMzNDU0NCwidXNlcl9pZCI6IlBkbVR6dkxROVFVdXVPUllyQmg4a1dEZTFmeDEiLCJzdWIiOiJQZG1UenZMUTlRVXV1T1JZckJoOGtXRGUxZngxIiwiaWF0IjoxNzUzNDE5NDcwLCJleHAiOjE3NTM0MjMwNzAsImVtYWlsIjoicmd2dmFybWEwMDlAZ21haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImZpcmViYXNlIjp7ImlkZW50aXRpZXMiOnsiZW1haWwiOlsicmd2dmFybWEwMDlAZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoicGFzc3dvcmQifX0.M-2cd1C9TaeX77COs7uFuaODXGX9yAbu3OPpsa-4L_xB87SAiLDRVGh6vEJTWVoe10OiZgtmUJbU0Cb6tTS-CSGICMJURHR3jEkcGcpWeF418Xb0Sctr3ldLSK_u86aZtuYawobhKRKPct_TksPpDkY9EKwAi0VsC8jNrY8okUSuqx8nHEUwnDxRIYJErxmsS6hN91LUnGEsZyoHG_LMUoyB3bdyxuGSJEkn7PDg3CFLUvmyfyMchzm5T2mHoPZMTQ47Dn38CDFy4mb5GXkWAZ7JqwLxK8MlBDpkt2e2ykePIJ_dhE3qnxgfhC-TFDCqhMXdGxayd6hCSESGi9bdGA" \
        -d "{\"className\":\"CLS10\",\"subjectName\":\"$subject\"}"
    echo "✅ Inserted subject: CLS10 -> $subject"
done

# Insert subjects for CLS11-MPC
for subject in PHYSICS MATHS1A MATHS1B CHEMISTRY EAMCET JEEMAINS JEEADV; do
    curl -X POST "$API_URL/v2/subject/insert" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjZkZTQwZjA0ODgxYzZhMDE2MTFlYjI4NGE0Yzk1YTI1MWU5MTEyNTAiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vZ290aGlua2Vyc3RlY2giLCJhdWQiOiJnb3RoaW5rZXJzdGVjaCIsImF1dGhfdGltZSI6MTc1MzMzNDU0NCwidXNlcl9pZCI6IlBkbVR6dkxROVFVdXVPUllyQmg4a1dEZTFmeDEiLCJzdWIiOiJQZG1UenZMUTlRVXV1T1JZckJoOGtXRGUxZngxIiwiaWF0IjoxNzUzNDE5NDcwLCJleHAiOjE3NTM0MjMwNzAsImVtYWlsIjoicmd2dmFybWEwMDlAZ21haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImZpcmViYXNlIjp7ImlkZW50aXRpZXMiOnsiZW1haWwiOlsicmd2dmFybWEwMDlAZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoicGFzc3dvcmQifX0.M-2cd1C9TaeX77COs7uFuaODXGX9yAbu3OPpsa-4L_xB87SAiLDRVGh6vEJTWVoe10OiZgtmUJbU0Cb6tTS-CSGICMJURHR3jEkcGcpWeF418Xb0Sctr3ldLSK_u86aZtuYawobhKRKPct_TksPpDkY9EKwAi0VsC8jNrY8okUSuqx8nHEUwnDxRIYJErxmsS6hN91LUnGEsZyoHG_LMUoyB3bdyxuGSJEkn7PDg3CFLUvmyfyMchzm5T2mHoPZMTQ47Dn38CDFy4mb5GXkWAZ7JqwLxK8MlBDpkt2e2ykePIJ_dhE3qnxgfhC-TFDCqhMXdGxayd6hCSESGi9bdGA" \
        -d "{\"className\":\"CLS11-MPC\",\"subjectName\":\"$subject\"}"
    echo "✅ Inserted subject: CLS11-MPC -> $subject"
done

# Insert subjects for CLS12-MPC
for subject in PHYSICS MATHS2A MATHS2B CHEMISTRY EAMCET JEEMAINS JEEADV; do
    curl -X POST "$API_URL/v2/subject/insert" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjZkZTQwZjA0ODgxYzZhMDE2MTFlYjI4NGE0Yzk1YTI1MWU5MTEyNTAiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vZ290aGlua2Vyc3RlY2giLCJhdWQiOiJnb3RoaW5rZXJzdGVjaCIsImF1dGhfdGltZSI6MTc1MzMzNDU0NCwidXNlcl9pZCI6IlBkbVR6dkxROVFVdXVPUllyQmg4a1dEZTFmeDEiLCJzdWIiOiJQZG1UenZMUTlRVXV1T1JZckJoOGtXRGUxZngxIiwiaWF0IjoxNzUzNDE5NDcwLCJleHAiOjE3NTM0MjMwNzAsImVtYWlsIjoicmd2dmFybWEwMDlAZ21haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImZpcmViYXNlIjp7ImlkZW50aXRpZXMiOnsiZW1haWwiOlsicmd2dmFybWEwMDlAZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoicGFzc3dvcmQifX0.M-2cd1C9TaeX77COs7uFuaODXGX9yAbu3OPpsa-4L_xB87SAiLDRVGh6vEJTWVoe10OiZgtmUJbU0Cb6tTS-CSGICMJURHR3jEkcGcpWeF418Xb0Sctr3ldLSK_u86aZtuYawobhKRKPct_TksPpDkY9EKwAi0VsC8jNrY8okUSuqx8nHEUwnDxRIYJErxmsS6hN91LUnGEsZyoHG_LMUoyB3bdyxuGSJEkn7PDg3CFLUvmyfyMchzm5T2mHoPZMTQ47Dn38CDFy4mb5GXkWAZ7JqwLxK8MlBDpkt2e2ykePIJ_dhE3qnxgfhC-TFDCqhMXdGxayd6hCSESGi9bdGA" \
        -d "{\"className\":\"CLS12-MPC\",\"subjectName\":\"$subject\"}"
    echo "✅ Inserted subject: CLS12-MPC -> $subject"
done

# Insert subjects for CLS11-BIPC
for subject in PHYSICS BOTANY ZOOLOGY CHEMISTRY EAPCET NEET; do
    curl -X POST "$API_URL/v2/subject/insert" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjZkZTQwZjA0ODgxYzZhMDE2MTFlYjI4NGE0Yzk1YTI1MWU5MTEyNTAiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vZ290aGlua2Vyc3RlY2giLCJhdWQiOiJnb3RoaW5rZXJzdGVjaCIsImF1dGhfdGltZSI6MTc1MzMzNDU0NCwidXNlcl9pZCI6IlBkbVR6dkxROVFVdXVPUllyQmg4a1dEZTFmeDEiLCJzdWIiOiJQZG1UenZMUTlRVXV1T1JZckJoOGtXRGUxZngxIiwiaWF0IjoxNzUzNDE5NDcwLCJleHAiOjE3NTM0MjMwNzAsImVtYWlsIjoicmd2dmFybWEwMDlAZ21haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImZpcmViYXNlIjp7ImlkZW50aXRpZXMiOnsiZW1haWwiOlsicmd2dmFybWEwMDlAZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoicGFzc3dvcmQifX0.M-2cd1C9TaeX77COs7uFuaODXGX9yAbu3OPpsa-4L_xB87SAiLDRVGh6vEJTWVoe10OiZgtmUJbU0Cb6tTS-CSGICMJURHR3jEkcGcpWeF418Xb0Sctr3ldLSK_u86aZtuYawobhKRKPct_TksPpDkY9EKwAi0VsC8jNrY8okUSuqx8nHEUwnDxRIYJErxmsS6hN91LUnGEsZyoHG_LMUoyB3bdyxuGSJEkn7PDg3CFLUvmyfyMchzm5T2mHoPZMTQ47Dn38CDFy4mb5GXkWAZ7JqwLxK8MlBDpkt2e2ykePIJ_dhE3qnxgfhC-TFDCqhMXdGxayd6hCSESGi9bdGA" \
        -d "{\"className\":\"CLS11-BIPC\",\"subjectName\":\"$subject\"}"
    echo "✅ Inserted subject: CLS11-BIPC -> $subject"
done

# Insert subjects for CLS12-BIPC
for subject in PHYSICS BOTANY ZOOLOGY CHEMISTRY EAPCET NEET; do
    curl -X POST "$API_URL/v2/subject/insert" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjZkZTQwZjA0ODgxYzZhMDE2MTFlYjI4NGE0Yzk1YTI1MWU5MTEyNTAiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vZ290aGlua2Vyc3RlY2giLCJhdWQiOiJnb3RoaW5rZXJzdGVjaCIsImF1dGhfdGltZSI6MTc1MzMzNDU0NCwidXNlcl9pZCI6IlBkbVR6dkxROVFVdXVPUllyQmg4a1dEZTFmeDEiLCJzdWIiOiJQZG1UenZMUTlRVXV1T1JZckJoOGtXRGUxZngxIiwiaWF0IjoxNzUzNDE5NDcwLCJleHAiOjE3NTM0MjMwNzAsImVtYWlsIjoicmd2dmFybWEwMDlAZ21haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImZpcmViYXNlIjp7ImlkZW50aXRpZXMiOnsiZW1haWwiOlsicmd2dmFybWEwMDlAZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoicGFzc3dvcmQifX0.M-2cd1C9TaeX77COs7uFuaODXGX9yAbu3OPpsa-4L_xB87SAiLDRVGh6vEJTWVoe10OiZgtmUJbU0Cb6tTS-CSGICMJURHR3jEkcGcpWeF418Xb0Sctr3ldLSK_u86aZtuYawobhKRKPct_TksPpDkY9EKwAi0VsC8jNrY8okUSuqx8nHEUwnDxRIYJErxmsS6hN91LUnGEsZyoHG_LMUoyB3bdyxuGSJEkn7PDg3CFLUvmyfyMchzm5T2mHoPZMTQ47Dn38CDFy4mb5GXkWAZ7JqwLxK8MlBDpkt2e2ykePIJ_dhE3qnxgfhC-TFDCqhMXdGxayd6hCSESGi9bdGA" \
        -d "{\"className\":\"CLS12-BIPC\",\"subjectName\":\"$subject\"}"
    echo "✅ Inserted subject: CLS12-BIPC -> $subject"
done

echo "🎉 Class data initialization complete!"
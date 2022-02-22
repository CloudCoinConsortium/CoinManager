#!/bin/bash


x='{
        "cloudcoin": [{
                "nn":"1",
                "sn":"1280910",
                "an":["33f726eb987f9c60c47dc96768c8f74c","7332b296fea1b99180986cbffdcd1165","40fada67786fac9af0827aea456a7a8f","882c0fc7fec1588fa6086095e0dd553e","c1ea3e27ef772ac4cff014dce1b29444",
"2a456d32032c45e628365cc602f28927","97aa877ff89ab9c676c8cfff1db0932d","29e6185411381e4beb7c8181750617b5","6c74e3370373eff434be22de35454e5d","a5ec9b39ffaee3a22dd2e9b66657a303",
"71408f844a47e5af8fd1b91bf37b6937","7b41f21dec55d91bbbd394e8bda7029c","17875bb257ef068f6adf2da764674d97","0068def964e27c6de3a26fab62daffab","caba363742f31674a8ee4c2db5fb2a56",
"3ed522b4fcaf20de59c42ac4e1197084","705c395f03b1bff4e1a78c83f9523989","c05639a251a2dbc2bb7c23d326e79724","394458a8b90ae8328073265ffc4ca8a4","a06e3fd4f4598438f1cf7a3e3eb5ecff",
"b25024aca893e66a032992c7e1587df9","9be54436dba2b318b395a321633dbd18","9ca71de172aac5f51834ab291ed3bf75","d50371592d5efdedb51cb05c4a56b439","3a791950f2f1f3d797afb5d0bac81dbc"],
                "ed" : "10-2025",
                "pown": "ppppppppppppppppppppupppp",
                "aoid": []
        }, {
                "nn":"1",
                "sn":"1280914",
                "an":["3e46a7faf5182acdd135c22759b60f70","9fc967101831e5c1ac4645f21798fc3c","d9fe394d40e093650830a0c2ce29a42c","c0200fce0b70a14917be6155ddd9d160","a44c025bf221987c0c26702e44a2ff21",
"7c185698972db4f1cb441cd8b1730975","06185524394759bf57ee1966ff04878c","0cf5b33ab1f801368feb91c4f3cbc548","3b174c9c2fc6fa299e444639273b16f9","db71710fd0eb49ce559f66b4274689ab",
"d8507979f4b86fb35ba61a67e7679ef4","b69a4e3296fee9d041d415b7d910e384","ca1476324c36c87f220ef7e43d6be6f2","96b74105f38270f76aca593aa0e2bce0","d543b74caa40024988d403194619d322",
"4954197f1fb60ef481240956ea7d07b3","cceba2674fedc43efd5f82b6835b2560","e11b3abed2b35a73bbae0d06ab4a1c1e","993030f3bdc7d335d1eb40d134700748","dc0da33331c2f5ca58379bc4e8904d77",
"ba38f2b823281ffa1b5343e6478d15c7","58cbcb90ae575a62bf14e24eb5b72fc2","8c87836c6fba4c80ea9626695496564a","a47384b785e916799302a096cacb854a","ce4d16425ea8e27cd433b666b0d56d0a"],
                "ed" : "10-2025",
                "pown": "ppppppppppppppppppppupppp",
                "aoid": []
        }, {
                "nn":1,
                "sn":1280915,
                "an":["d8a2b50aa2ce06ebf05099ba32f2d1f8","d0e99482f5115adf1d0d21538a25f563","feab5db2149c86ac3751c7181f4541e7","15d8d28135e87e94056dc13bdbedfda2","6bf43614e74c9d2a80369f54306aceed",
"a39b7ca7cd340cb21dcf2061ce84f223","51a9235eb6b90c9daed912c51ee3ee45","65e9dff1886177765a561d9695cc907f","a5efe6a0ebc9ac92063d0116abfe54e8","33f310f99a8a09cc4bf97bf05892b5a9",
"7225414e04050b7a756bb8e890b73bcb","0a5673d5b3236db1fb2092e04bb87880","4fe11bb85233207ae1e688eda9c86d59","125630f62bf7cd896b6f573c076c18fd","963f47429ffc2104e95f62852a2ad211",
"c1537e0c7f6e00f2ee53a07949df6e6d","e1a3b47f1552425e06d23ea455395118","590e798ba8b0402be9ec92c9e87b80b9","e26dec8be4f0836aaf99be90190231df","be46fa4357eaede626f0cc692e6f383f",
"13819d87279eae762ba73977513fae76","e5c1b259ebd8a484238e63715d9e9a37","0f15ac8d35ba6e2bad67eb92f621c8b0","4c6f50a3c874abc6359aff58506c6942","d5d2946741cc1495b747a6eada971281"],
                "ed" : "10-2025",
                "pown": "ppppppppppppppppppppupppp",
                "aoid": []
        }]
}'


b=$(echo $x |base64 -w0)


#b=$(cat /home/alexander/axx.skywallet.cc.png |base64 -w0)

curl -v 'http://localhost:8888/api/v1/unpack' -d "{\"data\":\"$b\"}"
#curl -v 'http://localhost:8888/api/v1/unpack' -d @<(echo "{\"data\":\"$b\"}")

